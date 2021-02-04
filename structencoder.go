package jingo

// structencoder.go manages StructEncoder and its responsibilities.
// The general goal of the approach is to do as much of the necessary work as possible inside
// the 'compile' stage upon instantiation. This includes any logic, type assertions, buffering
// or otherwise. Changes made should consider first their ns/op impact and then their allocation
// profile also. Allocations should essentially remain at zero - albeit with the exclusion of the
// `.String()` stringer functionality which is somewhat out of our control.

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

// StructEncoder stores a set of instructions for converting a struct to a json document. It's
// useless to create an instance of this outside of `NewStructEncoder`.
type StructEncoder struct {
	instructions []func(t unsafe.Pointer, w *Buffer) // the instructionset to be executed during Marshal
	f            reflect.StructField                 // current field
	t            interface{}                         // type
	i            int                                 // iter
	cb           Buffer                              // side buffer for static data
	cpos         int                                 // side buffer position
}

// Marshal executes the instructions for a given type and writes the resulting
// json document to the io.Writer provided
func (e *StructEncoder) Marshal(s interface{}, w *Buffer) {

	p := unsafe.Pointer(reflect.ValueOf(s).Pointer())
	for i, l := 0, len(e.instructions); i < l; i++ {
		e.instructions[i](p, w)
	}
}

// NewStructEncoder compiles a set of instructions for marhsaling a struct shape to a JSON document.
func NewStructEncoder(t interface{}) *StructEncoder {
	e := &StructEncoder{}
	e.t = t
	tt := reflect.TypeOf(t)

	e.chunk("{")

	emit := 0 // track number of fields we emit
	// pass over each field in the struct to build up our instruction set for each
	for e.i = 0; e.i < tt.NumField(); e.i++ {
		e.f = tt.Field(e.i)

		tag, opts := parseTag(e.f.Tag.Get("json")) // we're using tags to nominate inclusion
		if tag == "" {
			continue
		}
		emit++

		// write the key
		if emit > 1 {
			e.chunk(",")
		}
		e.chunk(`"` + tag + `":`)

		switch {
		/// support calling .String() when the 'stringer' option is passed
		case opts.Contains("stringer") && reflect.ValueOf(e.t).Field(e.i).MethodByName("String").Kind() != reflect.Invalid:
			e.optInstrStringer()

		/// support calling .JSONEncode(*Buffer) when the 'encoder' option is passed
		case opts.Contains("encoder"):
			e.optInstrEncoder()

		/// support writing byteslice-like items using 'raw' option.
		case opts.Contains("raw"):
			e.optInstrRaw()

		/// suport escaping reserved json characters from byteslice-like items and slices
		case opts.Contains("escape"):
			e.optInstrEscape()

		/// time is a type of struct, not a kind, so somewhat of a special case here.
		case e.f.Type == timeType:
			e.val(ptrTimeToBuf)
		case e.f.Type.Kind() == reflect.Ptr && timeType == reflect.TypeOf(e.t).Field(e.i).Type.Elem():
			e.ptrval(ptrTimeToBuf)

		// write the value instruction depending on type
		case e.f.Type.Kind() == reflect.Ptr:
			// create an instruction which can read from a pointer field
			e.valueInst(e.f.Type.Elem().Kind(), e.ptrval)

		default:
			// create an instruction which reads from a standard field
			e.valueInst(e.f.Type.Kind(), e.val)
		}
	}

	e.chunk("}")
	e.flunk()

	return e
}

func (e *StructEncoder) optInstrStringer() {
	e.chunk(`"`)

	t := reflect.ValueOf(e.t).Field(e.i).Type()
	if e.f.Type.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	conv := func(v unsafe.Pointer, w *Buffer) {
		e, ok := reflect.NewAt(t, v).Interface().(fmt.Stringer)
		if !ok {
			return
		}
		w.WriteString(e.String())
	}

	if e.f.Type.Kind() == reflect.Ptr {
		e.ptrval(conv)
	} else {
		e.val(conv)
	}

	e.chunk(`"`)
}

func (e *StructEncoder) optInstrEncoder() {
	t := reflect.ValueOf(e.t).Field(e.i).Type()
	if e.f.Type.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	conv := func(v unsafe.Pointer, w *Buffer) {
		e, ok := reflect.NewAt(t, v).Interface().(JSONEncoder)
		if !ok {
			w.Write(null)
			return
		}
		e.JSONEncode(w)
	}

	if e.f.Type.Kind() == reflect.Ptr {
		e.ptrval(conv)
	} else {
		e.val(conv)
	}
}

func (e *StructEncoder) optInstrRaw() {
	conv := func(v unsafe.Pointer, w *Buffer) {
		s := *(*string)(v)
		if len(s) == 0 {
			w.Write(null)
			return
		}
		w.WriteString(s)
	}

	if e.f.Type.Kind() == reflect.Ptr {
		e.ptrval(conv)
	} else {
		e.val(conv)
	}
}

func (e *StructEncoder) optInstrEscape() {
	if e.f.Type.Kind() == reflect.Slice {
		e.flunk()

		/// create an escape string encoder internally instead of mirroring the struct, so people only need to pass the ,escape opt instead
		enc := NewSliceEncoder([]EscapeString{})
		f := e.f
		e.instructions = append(e.instructions, func(v unsafe.Pointer, w *Buffer) {
			var em interface{} = unsafe.Pointer(uintptr(v) + f.Offset)
			enc.Marshal(em, w)
		})
		return
	}

	if e.f.Type.Kind() == reflect.Ptr {
		e.ptrstringval(ptrEscapeStringToBuf)
	} else {
		e.chunk(`"`)
		e.val(ptrEscapeStringToBuf)
		e.chunk(`"`)
	}
}

// chunk writes a chunk of body data to the chunk buffer. only for writing static
//  structure and not dynamic values.
func (e *StructEncoder) chunk(b string) {
	e.cb.Write([]byte(b))
}

// flunk flushes whatever chunk data we've got buffered into a single instruction
func (e *StructEncoder) flunk() {

	b := e.cb.Bytes
	bs := b[e.cpos:]
	e.cpos = len(b)

	if len(bs) == 0 {
		return
	}

	e.instructions = append(e.instructions, func(_ unsafe.Pointer, w *Buffer) {
		w.Write(bs)
	})
}

/// valueInst works out the conversion function we need for `k` and creates an instruction to write it to the buffer
func (e *StructEncoder) valueInst(k reflect.Kind, instr func(func(unsafe.Pointer, *Buffer))) {

	switch k {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64:
		/// standard print
		conv, ok := typeconv[k]
		if !ok {
			return
		}
		instr(conv)

	case reflect.Array:
		/// support for primitives in arrays (proabbly need arrayencoder.go here if we want to take this further)
		e.chunk("[")

		conv, ok := typeconv[e.f.Type.Elem().Kind()]
		if !ok {
			return
		}

		offset := e.f.Type.Elem().Size()
		for i := 0; i < e.f.Type.Len(); i++ {
			if i > 0 {
				e.chunk(", ")
			}

			e.flunk()
			f := e.f
			i := i
			e.instructions = append(e.instructions, func(v unsafe.Pointer, w *Buffer) {
				conv(unsafe.Pointer(uintptr(v)+f.Offset+(uintptr(i)*offset)), w)
			})
		}

		e.chunk("]")

	case reflect.Slice:

		e.flunk()

		enc := NewSliceEncoder(reflect.ValueOf(e.t).Field(e.i).Interface())
		f := e.f
		e.instructions = append(e.instructions, func(v unsafe.Pointer, w *Buffer) {
			var em interface{} = unsafe.Pointer(uintptr(v) + f.Offset)
			enc.Marshal(em, w)
		})

	case reflect.String:

		/// for strings to be nullable they need a special instruction to write quotes conditionally.
		if e.f.Type.Kind() == reflect.Ptr {
			e.ptrstringval(ptrStringToBuf)
			return
		}

		// otherwise a standard quoted print instruction
		e.chunk(`"`)
		instr(ptrStringToBuf)
		e.chunk(`"`)

	case reflect.Struct:
		// create an instruction for the field name (as per val)
		e.flunk()

		if e.f.Type.Kind() == reflect.Ptr {

			/// now cater for it being a pointer to a struct
			var inf = reflect.New(reflect.TypeOf(e.t).Field(e.i).Type.Elem()).Elem().Interface()
			enc := NewStructEncoder(inf)
			// now create an instruction to marshal the field
			f := e.f
			e.instructions = append(e.instructions, func(v unsafe.Pointer, w *Buffer) {
				var em interface{} = unsafe.Pointer(*(*unsafe.Pointer)(unsafe.Pointer(uintptr(v) + f.Offset)))
				if em == unsafe.Pointer(nil) {
					w.Write(null)
					return
				}
				enc.Marshal(em, w)
			})
			return
		}

		// build a new StructEncoder for the type
		enc := NewStructEncoder(reflect.ValueOf(e.t).Field(e.i).Interface())
		// now create another instruction which calls marshal on the struct, passing our writer
		f := e.f
		e.instructions = append(e.instructions, func(v unsafe.Pointer, w *Buffer) {
			var em interface{} = unsafe.Pointer(uintptr(v) + f.Offset)
			enc.Marshal(em, w)
		})
		return

	case reflect.Invalid,
		reflect.Map,
		reflect.Interface,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Chan,
		reflect.Func,
		reflect.Uintptr,
		reflect.UnsafePointer:
		// no
		panic(fmt.Sprint("unsupported type ", e.f.Type.Kind(), e.f.Name))
	}
}

// val creates an instruction to read from a field we're marshaling
func (e *StructEncoder) val(conv func(unsafe.Pointer, *Buffer)) {

	e.flunk() // flush any chunk data we've buffered

	f := e.f
	e.instructions = append(e.instructions, func(v unsafe.Pointer, w *Buffer) {
		conv(unsafe.Pointer(uintptr(v)+f.Offset), w)
	})
}

// ptrval creates an instruction to read from a pointer field we're marshaling
func (e *StructEncoder) ptrval(conv func(unsafe.Pointer, *Buffer)) {

	e.flunk() // flush any chunk data we've buffered

	// avoids allocs at runtime
	null := []byte("null")

	f := e.f
	e.instructions = append(e.instructions, func(v unsafe.Pointer, w *Buffer) {

		p := unsafe.Pointer(*(*unsafe.Pointer)(unsafe.Pointer(uintptr(v) + f.Offset)))
		if p == unsafe.Pointer(nil) {
			w.Write(null)
			return
		}
		conv(p, w)
	})
}

// ptrstringval is essentially the same as ptrval but quotes strings if not nil
func (e *StructEncoder) ptrstringval(conv func(unsafe.Pointer, *Buffer)) {
	e.flunk() // flush any chunk data we've buffered

	// avoids allocs at runtime
	null := []byte("null")

	f := e.f
	e.instructions = append(e.instructions, func(v unsafe.Pointer, w *Buffer) {

		p := unsafe.Pointer(*(*unsafe.Pointer)(unsafe.Pointer(uintptr(v) + f.Offset)))
		if p == unsafe.Pointer(nil) {
			w.Write(null)
			return
		}

		// quotes need to be at runtime here because we don't know if we're going to have to null the field
		w.WriteByte('"')
		conv(p, w)
		w.WriteByte('"')
	})
}

// JSONEncoder works with the `.encoder` option. Fields can implement this to encode their own JSON string straight
// into the working buffer. This can be useful if you're working with interface fields at runtime.
type JSONEncoder interface {
	JSONEncode(*Buffer)
}

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
//
// this is jacked from the stdlib to remain compatible with that syntax.
type tagOptions string

// parseTag splits a struct field's json tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

var timeType = reflect.TypeOf(time.Time{})

// EscapeString can be used to cast your string slice encoders in replacement of `[]string` when using SliceEncoder directly.
// This is only necessary if you wish for the slice elements to be escaped of control sequences.
// e.g var mySliceEncoder = NewSliceEncoder([]jingo.EscapeString{})
// You can and should just use the `,escape` option on your struct fields when using StructEncoder.
type EscapeString string

var escapeStringType = reflect.TypeOf(EscapeString(""))
