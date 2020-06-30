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
	return NewSliceEncoderWithConfig(t, DefaultConfig())
}

// NewSliceEncoderWithConfig builds a new SliceEncoder using Config provided.
func NewSliceEncoderWithConfig(t interface{}, cfg Config) *SliceEncoder {
	e := &SliceEncoder{}

	e.tt = reflect.TypeOf(t)
	e.offset = e.tt.Elem().Size()

	if e.tt.Elem() == timeType {
		e.timeInstr()
		return e
	}

	// what type of encoding do we need
	switch e.tt.Elem().Kind() {
	case reflect.Slice:
		e.sliceInstr(cfg)

	case reflect.Struct:
		e.structInstr(cfg)

	case reflect.Map:
		e.mapInstr(cfg)

	case reflect.String:
		e.stringInstr()

	case reflect.Ptr:

		/// which pointer type
		if e.tt.Elem().Elem() == timeType {
			e.ptrTimeInstr()
			return e
		}

		switch e.tt.Elem().Elem().Kind() {
		case reflect.Slice:
			e.ptrSliceInstr(cfg)

		case reflect.Struct:
			e.ptrStrctInstr(cfg)

		case reflect.Map:
			e.ptrMapInstr(cfg)

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

// // avoid allocs in the instruction
var (
	null = []byte("null")
	zero = uintptr(0)
)

func (e *SliceEncoder) sliceInstr(cfg Config) {
	enc := NewSliceEncoderWithConfig(reflect.New(e.tt.Elem()).Elem().Interface(), cfg)
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}
			s := unsafe.Pointer(sl.Data + (i * e.offset))
			enc.Marshal(s, w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) structInstr(cfg Config) {
	enc := NewStructEncoderWithConfig(reflect.New(e.tt.Elem()).Elem().Interface(), cfg)
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}
			s := unsafe.Pointer(sl.Data + (i * e.offset))
			enc.Marshal(s, w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) mapInstr(cfg Config) {
	enc := NewMapEncoderWithConfig(reflect.New(e.tt.Elem()).Elem().Interface(), cfg)
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}
			s := unsafe.Pointer(uintptr(sl.Data) + (i * e.offset))
			enc.Marshal(s, w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) stringInstr() {
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {

			if i == 0 {
				w.WriteByte('"')
			}

			if i > zero {
				w.Write([]byte(`","`))
			}

			ptrStringToBuf(unsafe.Pointer(sl.Data+(i*e.offset)), w)

			if i == uintptr(sl.Len)-1 {
				w.WriteByte('"')
			}
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) otherInstr() {

	conv, ok := typeconv[e.tt.Elem().Kind()]
	if !ok {
		return
	}

	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}
			conv(unsafe.Pointer(sl.Data+(i*e.offset)), w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) timeInstr() {
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}
			w.WriteByte('"')
			ptrTimeToBuf(unsafe.Pointer(sl.Data+(i*e.offset)), w)
			w.WriteByte('"')
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) ptrSliceInstr(cfg Config) {
	enc := NewSliceEncoderWithConfig(reflect.New(e.tt.Elem()).Elem().Elem().Interface(), cfg)
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*uintptr)(unsafe.Pointer(sl.Data + (i * e.offset))))
			if s == unsafe.Pointer(nil) {
				w.Write(null)
				continue
			}
			enc.Marshal(s, w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) ptrStrctInstr(cfg Config) {
	enc := NewStructEncoderWithConfig(reflect.New(e.tt.Elem().Elem()).Elem().Interface(), cfg)
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*uintptr)(unsafe.Pointer(sl.Data + (i * e.offset))))
			if s == unsafe.Pointer(nil) {
				w.Write(null)
				continue
			}
			enc.Marshal(s, w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) ptrMapInstr(cfg Config) {
	enc := NewMapEncoderWithConfig(reflect.New(e.tt.Elem().Elem()).Elem().Interface(), cfg)
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := *(*unsafe.Pointer)(unsafe.Pointer(uintptr(sl.Data) + (i * e.offset)))
			if s == unsafe.Pointer(nil) {
				w.Write(null)
				continue
			}
			enc.Marshal(s, w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) ptrStringInstr() {
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*uintptr)(unsafe.Pointer(sl.Data + (i * e.offset))))
			if s == unsafe.Pointer(nil) {
				w.Write(null)
				continue
			}
			w.WriteByte('"')
			ptrStringToBuf(s, w)
			w.WriteByte('"')
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) ptrOtherInstr() {

	conv, ok := typeconv[e.tt.Elem().Elem().Kind()]
	if !ok {
		return
	}

	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*uintptr)(unsafe.Pointer(sl.Data + (i * e.offset))))
			if s == unsafe.Pointer(nil) {
				w.Write(null)
				continue
			}
			conv(s, w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) ptrTimeInstr() {
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*reflect.SliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*uintptr)(unsafe.Pointer(sl.Data + (i * e.offset))))
			if s == unsafe.Pointer(nil) {
				w.Write(null)
				continue
			}
			w.WriteByte('"')
			ptrTimeToBuf(s, w)
			w.WriteByte('"')
		}

		w.WriteByte(']')
	}
}
