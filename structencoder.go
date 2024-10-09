package jingo

// structencoder.go manages StructEncoder and its responsibilities.
// The general goal of the approach is to do as much of the necessary work as possible inside
// the 'compile' stage upon instantiation. This includes any logic, type assertions, buffering
// or otherwise. Changes made should consider first their ns/op impact and then their allocation
// profile also. Allocations should essentially remain at zero - albeit with the exclusion of the
// `.String()` stringer functionality which is somewhat out of our control.

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

// instruction describes the different ways we can execute a single instruction at runtime.
// static, leapFun/offset, fun are mutually exclusive. we've used a concrete type for speed.
type instruction struct {
	static  []byte                        // provides a fast path for writing static chunks without needing an instruction function
	sub     []instruction                 // sub instructions for omitempty
	kind    int                           // used to switch special paths in Marshal, like string fast path
	offset  uintptr                       // used in conjunction with leapFun
	leapFun func(unsafe.Pointer, *Buffer) // provides a fast path for simple write & avoids wrapping function to capture offset
	fun     func(unsafe.Pointer, *Buffer) // full instruction function for when the approaches above fail
	isZero  func(unsafe.Pointer) bool     // isZero tests if pointer is zero
}

const (
	kindNormal = iota
	kindStringField
	kindStatic
	kindInt
	KindOmit
)

// iface describes the memory footprint of interface{}
type iface struct {
	Type, Data unsafe.Pointer
}

// StructEncoder stores a set of instructions for converting a struct to a json document. It's
// useless to create an instance of this outside of `NewStructEncoder`.
type StructEncoder struct {
	instructions []instruction       // the instructionset to be executed during Marshal
	f            reflect.StructField // current field
	t            interface{}         // type
	i            int                 // iter
	cb           Buffer              // side buffer for static data
	cpos         int                 // side buffer position
	o            bool                // field is omitempty
}

// Marshal executes the instructions for a given type and writes the resulting
// json document to the io.Writer provided
func (e *StructEncoder) Marshal(s interface{}, w *Buffer) {

	p := (*(*iface)(unsafe.Pointer(&s))).Data

	for i := 0; i < len(e.instructions); i++ {
		if e.instructions[i].kind == kindStatic { // static data fast path
			w.Write(e.instructions[i].static)
			continue
		} else if e.instructions[i].kind == kindStringField { // string fields fast path, allows inlining of whole write
			ptrStringToBuf(unsafe.Pointer(uintptr(p)+e.instructions[i].offset), w)
			continue
		} else if e.instructions[i].kind == kindInt { // int fields fast path, allows inlining of whole write
			ptrIntToBuf(unsafe.Pointer(uintptr(p)+e.instructions[i].offset), w)
			continue
		} else if e.instructions[i].leapFun != nil { // simple 'conv' function fast path
			e.instructions[i].leapFun(unsafe.Pointer(uintptr(p)+e.instructions[i].offset), w)
			continue
		} else if e.instructions[i].kind == KindOmit {
			// this is managed by omitFlunk
			w.Write(e.instructions[i].static)
			if e.instructions[i].isZero(unsafe.Pointer(uintptr(p) + e.instructions[i].offset)) {
				// this is managed by setupLastField
				if e.instructions[i].fun != nil {
					e.instructions[i].fun(nil, w)
				}
				continue
			}

			for si := 0; si < len(e.instructions[i].sub); si++ {
				if e.instructions[i].sub[si].kind == kindStatic { // static data fast path
					w.Write(e.instructions[i].sub[si].static)
					continue
				} else if e.instructions[i].sub[si].kind == kindStringField { // string fields fast path, allows inlining of whole write
					ptrStringToBuf(unsafe.Pointer(uintptr(p)+e.instructions[i].sub[si].offset), w)
					continue
				} else if e.instructions[i].sub[si].kind == kindInt { // int fields fast path, allows inlining of whole write
					ptrIntToBuf(unsafe.Pointer(uintptr(p)+e.instructions[i].sub[si].offset), w)
					continue
				} else if e.instructions[i].sub[si].leapFun != nil { // simple 'conv' function fast path
					e.instructions[i].sub[si].leapFun(unsafe.Pointer(uintptr(p)+e.instructions[i].sub[si].offset), w)
					continue
				}

				e.instructions[i].sub[si].fun(p, w) // all other instruction types
			}

			continue
		}

		e.instructions[i].fun(p, w) // all other instruction types
	}
}

// NewStructEncoder compiles a set of instructions for marhsaling a struct shape to a JSON document.
func NewStructEncoder(t interface{}) *StructEncoder {
	e := &StructEncoder{}
	e.t = t
	tt := reflect.TypeOf(t)

	e.chunk("{")

	// pass over each field in the struct to build up our instruction set for each
	for e.i = 0; e.i < tt.NumField(); e.i++ {
		e.f = tt.Field(e.i)

		tag, opts := parseTag(e.f.Tag.Get("json")) // we're using tags to nominate inclusion
		if tag == "" {
			continue
		}

		e.o = false
		e.o = opts.Contains("omitempty") && e.f.Type.Kind() != reflect.Struct
		if e.o {
			e.omitFlunk()
		}

		e.chunk(`"` + tag + `":`)

		switch {
		/// support calling .String() when the 'stringer' option is passed
		case opts.Contains("stringer") && reflect.ValueOf(e.t).Field(e.i).MethodByName("String").Kind() != reflect.Invalid:
			e.optInstrStringer()

		/// support calling .JSONEncode(*Buffer) when the 'encoder' option is passed
		case opts.Contains("encoder"):

			// requrie explicit opt-in for JSONMarshaler implementation
			t := reflect.ValueOf(e.t).Field(e.i).Type()
			if t.Kind() != reflect.Ptr {
				t = reflect.PtrTo(t)
			}

			if _, ok := t.MethodByName("EncodeJSON"); ok {
				e.optInstrEncoderWriter()
				break
			}

			// default to JSONEncoder implementation for any other encoder fields
			e.optInstrEncoder()

		/// support writing byteslice-like items using 'raw' option.
		case opts.Contains("raw"):
			e.optInstrRaw()

		/// suport escaping reserved json characters from byteslice-like items and slices
		case opts.Contains("escape"):
			e.optInstrEscape()

		/// time is a type of struct, not a kind, so somewhat of a special case here.
		case e.f.Type == timeType:
			e.chunk(`"`)
			e.val(ptrTimeToBuf)
			e.chunk(`"`)
		case e.f.Type.Kind() == reflect.Ptr && timeType == reflect.TypeOf(e.t).Field(e.i).Type.Elem():
			e.ptrstringval(ptrTimeToBuf)

		// write the value instruction depending on type
		case e.f.Type.Kind() == reflect.Ptr:
			// create an instruction which can read from a pointer field
			e.valueInst(e.f.Type.Elem().Kind(), e.ptrval)

		default:
			// create an instruction which reads from a standard field
			e.valueInst(e.f.Type.Kind(), e.val)
		}

		e.chunk(",")

		// if omitempty we need to flunk so just this field is omitted
		if e.o {
			e.flunk()
		}
	}

	e.setupLastField() // for comma handling
	e.chunk("}")

	e.flunk()

	return e
}

// setupLastField manually removes the comma from the last field.
//
// If it is an omit instruction we know to remove the last byte always.
// We then add a function to the general fun field (that never gets used) to reset the comma if this field is omitted
//
// If it's not omitempty we check if the last item is a comma as it could be a '{' and remove the comma.
func (e *StructEncoder) setupLastField() {
	if e.o {
		instr := &e.instructions[len(e.instructions)-1]
		subInstr := &instr.sub[len(instr.sub)-1]
		subInstr.static = subInstr.static[:len(subInstr.static)-1]

		instr.fun = func(_ unsafe.Pointer, w *Buffer) {
			if w.Bytes[len(w.Bytes)-1] == ',' { // remove any superfluous "," in the event the last field was omitted
				w.Bytes = w.Bytes[:len(w.Bytes)-1]
			}
		}
		e.o = false

	} else if e.cb.Bytes[len(e.cb.Bytes)-1] == ',' {
		// it doesn't matter if this makes the slice empty as this is before the flunk
		e.cb.Bytes = e.cb.Bytes[:len(e.cb.Bytes)-1]
	}
}

// appendInstruction is used to add an instruction to the decoder, we use this to ensure that omitEmpty is being set regardless of what instruction type is being passed through.
// If it is not an omit empty, it just appends it into the instructions list
//
// This also manages the omitempty functions on the parent instruction that allows us to perform isZero checks.
func (e *StructEncoder) appendInstruction(instr instruction) {
	// if this is an omit field we append to the current instruction instead of appending a new one for the isZero check
	if e.o {
		omitInstr := &e.instructions[len(e.instructions)-1]
		// if it is not a static kind, we know to also append the isZero check and offset
		if instr.kind != kindStatic {
			omitInstr.isZero = getIsZeroFunc[e.f.Type.Kind()]
			omitInstr.offset = instr.offset
		}

		omitInstr.sub = append(omitInstr.sub, instr)
		return
	}

	e.instructions = append(e.instructions, instr)
}

func (e *StructEncoder) appendInstructionFun(fun func(unsafe.Pointer, *Buffer), offset uintptr) {
	e.appendInstruction(instruction{fun: fun, offset: offset})
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

func (e *StructEncoder) optInstrEncoderWriter() {
	t := reflect.ValueOf(e.t).Field(e.i).Type()
	if e.f.Type.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	conv := func(v unsafe.Pointer, w *Buffer) {
		e, ok := reflect.NewAt(t, v).Interface().(JSONMarshaler)
		if !ok {
			w.Write(null)
			return
		}

		e.EncodeJSON(w)
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

		e.appendInstructionFun(func(v unsafe.Pointer, w *Buffer) {
			var em interface{} = unsafe.Pointer(uintptr(v) + f.Offset)
			enc.Marshal(em, w)
		}, f.Offset)
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
//
//	structure and not dynamic values.
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

	e.appendInstruction(instruction{static: bs, kind: kindStatic})
}

// omitFlunk flushes whatever chunk data we've got buffered into a single instruction and exits/enters into omitMarshal
// this is to ensure that everything that needs to be written out is already complete before the isZero check
func (e *StructEncoder) omitFlunk() {
	b := e.cb.Bytes
	bs := b[e.cpos:]
	e.cpos = len(b)

	if len(bs) == 0 {
		e.instructions = append(e.instructions, instruction{kind: KindOmit})
		return
	}

	e.instructions = append(e.instructions, instruction{static: bs, kind: KindOmit})
}

// valueInst works out the conversion function we need for `k` and creates an instruction to write it to the buffer
func (e *StructEncoder) valueInst(k reflect.Kind, instr func(func(unsafe.Pointer, *Buffer))) {

	switch k {

	case reflect.Int:

		/// fast path for int fields
		if e.f.Type.Kind() == reflect.Ptr {
			instr(ptrIntToBuf)
			return
		}
		e.flunk()
		e.appendInstruction(instruction{offset: e.f.Offset, kind: kindInt})

	case reflect.Bool,
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
			e.appendInstructionFun(func(v unsafe.Pointer, w *Buffer) {
				conv(unsafe.Pointer(uintptr(v)+f.Offset+(uintptr(i)*offset)), w)
			}, f.Offset)
		}

		e.chunk("]")

	case reflect.Slice:

		e.flunk()

		enc := NewSliceEncoder(reflect.ValueOf(e.t).Field(e.i).Interface())
		f := e.f
		e.appendInstructionFun(func(v unsafe.Pointer, w *Buffer) {
			enc.Marshal(unsafe.Pointer(uintptr(v)+f.Offset), w)
		}, f.Offset)

	case reflect.String:

		/// for strings to be nullable they need a special instruction to write quotes conditionally.
		if e.f.Type.Kind() == reflect.Ptr {
			e.ptrstringval(ptrStringToBuf)
			return
		}

		// otherwise a standard quoted print instruction
		e.chunk(`"`)

		/// fast path for strings
		e.flunk() // flush any chunk data we've buffered
		e.appendInstruction(instruction{offset: e.f.Offset, kind: kindStringField})
		e.chunk(`"`)

	case reflect.Struct:
		// create an instruction for the field name (as per val)
		e.flunk()

		if e.f.Type.Kind() == reflect.Ptr {

			/// now cater for it being a pointer to a struct
			var inf = reflect.New(reflect.TypeOf(e.t).Field(e.i).Type.Elem()).Elem().Interface()

			var enc *StructEncoder
			if e.t == inf {
				// handle recursive structs by re-using the current encoder
				enc = e
			} else {
				enc = NewStructEncoder(inf)
			}

			// now create an instruction to marshal the field
			f := e.f
			e.appendInstructionFun(func(v unsafe.Pointer, w *Buffer) {
				var em interface{} = unsafe.Pointer(*(*unsafe.Pointer)(unsafe.Pointer(uintptr(v) + f.Offset)))
				if em == unsafe.Pointer(nil) {
					w.Write(null)
					return
				}

				enc.Marshal(em, w)
			}, f.Offset)
			return
		}

		// build a new StructEncoder for the type
		enc := NewStructEncoder(reflect.ValueOf(e.t).Field(e.i).Interface())
		// now create another instruction which calls marshal on the struct, passing our writer
		f := e.f
		e.appendInstructionFun(func(v unsafe.Pointer, w *Buffer) {
			enc.Marshal(unsafe.Pointer(uintptr(v)+f.Offset), w)
		}, f.Offset)
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
	e.appendInstruction(instruction{leapFun: conv, offset: e.f.Offset})
}

// ptrval creates an instruction to read from a pointer field we're marshaling
func (e *StructEncoder) ptrval(conv func(unsafe.Pointer, *Buffer)) {

	e.flunk() // flush any chunk data we've buffered

	// avoids allocs at runtime
	null := []byte("null")

	f := e.f
	e.appendInstructionFun(func(v unsafe.Pointer, w *Buffer) {
		p := *(*unsafe.Pointer)(unsafe.Pointer(uintptr(v) + f.Offset))
		if p == unsafe.Pointer(nil) {
			w.Write(null)
			return
		}
		conv(p, w)
	}, f.Offset)
}

// ptrstringval is essentially the same as ptrval but quotes strings if not nil
func (e *StructEncoder) ptrstringval(conv func(unsafe.Pointer, *Buffer)) {
	e.flunk() // flush any chunk data we've buffered

	// avoids allocs at runtime
	null := []byte("null")

	f := e.f
	e.appendInstructionFun(func(v unsafe.Pointer, w *Buffer) {
		p := *(*unsafe.Pointer)(unsafe.Pointer(uintptr(v) + f.Offset))
		if p == unsafe.Pointer(nil) {
			w.Write(null)
			return
		}

		// quotes need to be at runtime here because we don't know if we're going to have to null the field
		w.WriteByte('"')
		conv(p, w)
		w.WriteByte('"')
	}, f.Offset)
}

// JSONEncoder works with the `.encoder` option. Fields can implement this to encode their own JSON string straight
// into the working buffer. This can be useful if you're working with interface fields at runtime.
type JSONEncoder interface {
	JSONEncode(*Buffer)
}

// JSONMarshaler works with the `.encoder` option. Fields can implement this to encode their own JSON string straight
// into the provided `io.Writer`. This is useful if you require the functionality of `JSONEncoder` but don't want the hard
// dependency on `Buffer`.
type JSONMarshaler interface {
	EncodeJSON(io.Writer)
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
