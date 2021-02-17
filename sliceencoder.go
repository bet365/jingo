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
	cfg         Config
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
	e := &SliceEncoder{cfg: cfg}

	e.tt = reflect.TypeOf(t)
	e.offset = e.tt.Elem().Size()

	// see if we can select based on a specific type
	switch e.tt.Elem() {
	case timeType:
		e.timeInstr()
		return e
	case escapeStringType:
		e.stringInstr(ptrEscapeStringToBuf)
		return e
	}

	// what type of encoding do we need
	switch e.tt.Elem().Kind() {
	case reflect.Slice:
		e.sliceInstr()

	case reflect.Struct:
		e.structInstr()

	case reflect.Map:
		e.mapInstr()

	case reflect.String:
		e.stringInstr(ptrStringToBuf)

	case reflect.Ptr:

		/// which pointer type
		switch e.tt.Elem().Elem() {
		case timeType:
			e.ptrTimeInstr()
			return e
		case escapeStringType:
			e.ptrStringInstr(ptrEscapeStringToBuf)
			return e
		}

		switch e.tt.Elem().Elem().Kind() {
		case reflect.Slice:
			e.ptrSliceInstr()

		case reflect.Struct:
			e.ptrStrctInstr()

		case reflect.Map:
			e.ptrMapInstr()

		case reflect.String:
			e.ptrStringInstr(ptrStringToBuf)

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

// sliceHeader is a replacement for reflect.SliceHeader which forces the uintptr conversion to be done inline to
// play nice with vet and the unsafe conversion rules
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func (e *SliceEncoder) sliceInstr() {
	enc := NewSliceEncoderWithConfig(reflect.New(e.tt.Elem()).Elem().Interface(), e.cfg)
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

func (e *SliceEncoder) structInstr() {
	enc := NewStructEncoderWithConfig(reflect.New(e.tt.Elem()).Elem().Interface(), e.cfg)
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

func (e *SliceEncoder) mapInstr() {
	enc := NewMapEncoderWithConfig(reflect.New(e.tt.Elem()).Elem().Interface(), e.cfg)
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

func (e *SliceEncoder) stringInstr(conv func(unsafe.Pointer, *Buffer)) {
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {

			if i == 0 {
				w.WriteByte('"')
			}

			if i > zero {
				w.Write([]byte(`","`))
			}

			conv(unsafe.Pointer(uintptr(sl.Data)+(i*e.offset)), w)

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

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}
			conv(unsafe.Pointer(uintptr(sl.Data)+(i*e.offset)), w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) timeInstr() {
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}
			w.WriteByte('"')
			ptrTimeToBuf(unsafe.Pointer(uintptr(sl.Data)+(i*e.offset)), w)
			w.WriteByte('"')
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) ptrSliceInstr() {
	enc := NewSliceEncoderWithConfig(reflect.New(e.tt.Elem()).Elem().Elem().Interface(), e.cfg)
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*unsafe.Pointer)(unsafe.Pointer(uintptr(sl.Data) + (i * e.offset))))
			if s == unsafe.Pointer(nil) {
				w.Write(null)
				continue
			}
			enc.Marshal(s, w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) ptrStrctInstr() {
	enc := NewStructEncoderWithConfig(reflect.New(e.tt.Elem().Elem()).Elem().Interface(), e.cfg)
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*unsafe.Pointer)(unsafe.Pointer(uintptr(sl.Data) + (i * e.offset))))
			if s == unsafe.Pointer(nil) {
				w.Write(null)
				continue
			}
			enc.Marshal(s, w)
		}

		w.WriteByte(']')
	}
}

func (e *SliceEncoder) ptrMapInstr() {
	enc := NewMapEncoderWithConfig(reflect.New(e.tt.Elem().Elem()).Elem().Interface(), e.cfg)
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

func (e *SliceEncoder) ptrStringInstr(conv func(unsafe.Pointer, *Buffer)) {
	e.instruction = func(v unsafe.Pointer, w *Buffer) {
		w.WriteByte('[')

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*unsafe.Pointer)(unsafe.Pointer(uintptr(sl.Data) + (i * e.offset))))
			if s == unsafe.Pointer(nil) {
				w.Write(null)
				continue
			}
			w.WriteByte('"')
			conv(s, w)
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

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*unsafe.Pointer)(unsafe.Pointer(uintptr(sl.Data) + (i * e.offset))))
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

		sl := *(*sliceHeader)(v)
		for i := uintptr(0); i < uintptr(sl.Len); i++ {
			if i > zero {
				w.WriteByte(',')
			}

			s := unsafe.Pointer(*(*unsafe.Pointer)(unsafe.Pointer(uintptr(sl.Data) + (i * e.offset))))
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
