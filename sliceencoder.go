package jingo

// sliceencoder.go manages SliceEncoder and its responsibilities.
// SliceEncoder follows the same principle of structencoder.go in that it generates lightweight
// instructions as part of its compile stage which are executed later during the Marshal. The
// slight difference here is that instruction is singular rather than plural. This is due to the
// need for slice iteration to be managed from the instruction itself (at 'runtime') as a result
// of slices being of variable length.

import (
	"reflect"
	"unsafe"
)

// SliceEncoder stores a set of instructions for building a JSON document from a slice at runtime.
type SliceEncoder struct {
	instruction func(t unsafe.Pointer, w *Buffer)
	tt          reflect.Type
	offset      uintptr
}

// Marshal executes the instruction set built up by NewSliceEncoder
func (e *SliceEncoder) Marshal(s interface{}, w *Buffer) {

	p := unsafe.Pointer(reflect.ValueOf(s).Pointer())
	e.instruction(p, w)
}

// NewSliceEncoder builds a new SliceEncoder
func NewSliceEncoder(t interface{}) *SliceEncoder {
	e := &SliceEncoder{}

	e.tt = reflect.TypeOf(t)
	e.offset = e.tt.Elem().Size()

	// what type of encoding do we need
	switch e.tt.Elem().Kind() {
	case reflect.Slice:
		e.sliceInstr()

	case reflect.Struct:
		e.structInstr()

	case reflect.String:
		e.stringInstr()

	case reflect.Ptr:
		/// which pointer type
		switch e.tt.Elem().Elem().Kind() {
		case reflect.Slice:
			e.ptrSliceInstr()

		case reflect.Struct:
			e.ptrStrctInstr()

		case reflect.String:
			e.ptrStringInstr()

		default:
			e.ptrOtherInstr()
		}

	default:
		e.otherInstr()
	}

	return e
}

// avoid allocs in the instruction
var (
	comma = []byte(",")
	open  = []byte("[")
	close = []byte("]")
	quote = []byte(`"`)
	zero  = uintptr(0)
)

func (e *SliceEncoder) sliceInstr() {
	enc := NewSliceEncoder(reflect.New(e.tt.Elem()).Elem().Interface())
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.Write(open)

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.Write(comma)
			}
			s := unsafe.Pointer(sl.Data + (i * e.offset))
			enc.Marshal(s, w)
		}

		w.Write(close)
	}
}

func (e *SliceEncoder) structInstr() {
	enc := NewStructEncoder(reflect.New(e.tt.Elem()).Elem().Interface())
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.Write(open)

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.Write(comma)
			}
			s := unsafe.Pointer(sl.Data + (i * e.offset))
			enc.Marshal(s, w)
		}

		w.Write(close)
	}
}

func (e *SliceEncoder) stringInstr() {
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.Write(open)

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.Write(comma)
			}
			w.Write(quote)
			ptrStringToBuf(unsafe.Pointer(sl.Data+(i*e.offset)), w)
			w.Write(quote)
		}

		w.Write(close)
	}
}

func (e *SliceEncoder) otherInstr() {

	conv, ok := typeconv[e.tt.Elem().Kind()]
	if !ok {
		return
	}

	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.Write(open)

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.Write(comma)
			}
			conv(unsafe.Pointer(sl.Data+(i*e.offset)), w)
		}

		w.Write(close)
	}
}

func (e *SliceEncoder) ptrSliceInstr() {
	enc := NewSliceEncoder(reflect.New(e.tt.Elem()).Elem().Elem().Interface())
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.Write(open)

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.Write(comma)
			}
			s := unsafe.Pointer(*(*uintptr)(unsafe.Pointer(sl.Data + (i * e.offset))))
			enc.Marshal(s, w)
		}

		w.Write(close)
	}
}

func (e *SliceEncoder) ptrStrctInstr() {
	enc := NewStructEncoder(reflect.New(e.tt.Elem().Elem()).Elem().Interface())
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.Write(open)

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.Write(comma)
			}
			s := unsafe.Pointer(*(*uintptr)(unsafe.Pointer(sl.Data + (i * e.offset))))
			enc.Marshal(s, w)
		}

		w.Write(close)
	}
}

func (e *SliceEncoder) ptrStringInstr() {
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.Write(open)

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.Write(comma)
			}
			w.Write(quote)
			ptrStringToBuf(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(sl.Data + (i * e.offset)))), w)
			w.Write(quote)
		}

		w.Write(close)
	}
}

func (e *SliceEncoder) ptrOtherInstr() {

	conv, ok := typeconv[e.tt.Elem().Elem().Kind()]
	if !ok {
		return
	}

	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.Write(open)

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.Write(comma)
			}
			conv(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(sl.Data + (i * e.offset)))), w)
		}

		w.Write(close)
	}
}
