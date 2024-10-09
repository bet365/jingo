package jingo

// ptrconvert.go declares a number of primitive form -> buffer conversion
// functions based on an unsafe.Pointer input. We're using the implementation
// from the standard library here, which don't perform badly but are a
// candidate for a more high performance implementation to be introduced.

import (
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

var typeconv = map[reflect.Kind]func(unsafe.Pointer, *Buffer){
	reflect.Bool:    ptrBoolToBuf,
	reflect.Int:     ptrIntToBuf,
	reflect.Int8:    ptrInt8ToBuf,
	reflect.Int16:   ptrInt16ToBuf,
	reflect.Int32:   ptrInt32ToBuf,
	reflect.Int64:   ptrInt64ToBuf,
	reflect.Uint:    ptrUintToBuf,
	reflect.Uint8:   ptrUint8ToBuf,
	reflect.Uint16:  ptrUint16ToBuf,
	reflect.Uint32:  ptrUint32ToBuf,
	reflect.Uint64:  ptrUint64ToBuf,
	reflect.Float32: ptrFloat32ToBuf,
	reflect.Float64: ptrFloat64ToBuf,
	reflect.String:  ptrStringToBuf,
}

var btrue, bfalse = []byte("true"), []byte("false")

func ptrBoolToBuf(v unsafe.Pointer, b *Buffer) {
	r := *(*bool)(v)
	if r {
		b.Write(btrue)
	} else {
		b.Write(bfalse)
	}
}

func ptrIntToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendInt(b.Bytes, int64(*(*int)(v)), 10)
}

func ptrInt8ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendInt(b.Bytes, int64(*(*int8)(v)), 10)
}

func ptrInt16ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendInt(b.Bytes, int64(*(*int16)(v)), 10)
}

func ptrInt32ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendInt(b.Bytes, int64(*(*int32)(v)), 10)
}

func ptrInt64ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendInt(b.Bytes, *(*int64)(v), 10)
}

func ptrUintToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendUint(b.Bytes, uint64(*(*uint)(v)), 10)
}

func ptrUint8ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendUint(b.Bytes, uint64(*(*uint8)(v)), 10)
}

func ptrUint16ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendUint(b.Bytes, uint64(*(*uint16)(v)), 10)
}

func ptrUint32ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendUint(b.Bytes, uint64(*(*uint32)(v)), 10)
}

func ptrUint64ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendUint(b.Bytes, *(*uint64)(v), 10)
}

func ptrFloat32ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendFloat(b.Bytes, float64(*(*float32)(v)), 'f', -1, 32)
}

func ptrFloat64ToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = strconv.AppendFloat(b.Bytes, *(*float64)(v), 'f', -1, 64)
}

//go:nocheckptr
func ptrStringToBuf(v unsafe.Pointer, b *Buffer) {
	b.WriteString(*(*string)(v))
}

func ptrTimeToBuf(v unsafe.Pointer, b *Buffer) {
	b.Bytes = (*time.Time)(v).AppendFormat(b.Bytes, time.RFC3339Nano)
}

func ptrEscapeStringToBuf(v unsafe.Pointer, w *Buffer) {
	bs := *(*string)(v)

	pos := 0
	for i := 0; i < len(bs); i++ {
		switch bs[i] {
		case '\\', '"':
			if pos < i {
				w.WriteString(bs[pos:i])
			}
			pos = i + 1

			w.WriteByte('\\')
			w.WriteByte(bs[i])
		case '\n':
			if pos < i {
				w.WriteString(bs[pos:i])
			}
			pos = i + 1

			w.WriteString(`\n`)
		case '\r':
			if pos < i {
				w.WriteString(bs[pos:i])
			}
			pos = i + 1

			w.WriteString(`\r`)
		case '\t':
			if pos < i {
				w.WriteString(bs[pos:i])
			}
			pos = i + 1

			w.WriteString(`\t`)
		}
	}

	if pos < len(bs) {
		w.WriteString(bs[pos:])
	}
}

var getIsZeroFunc = map[reflect.Kind]func(unsafe.Pointer) bool{
	reflect.Ptr:        func(v unsafe.Pointer) bool { return *(*unsafe.Pointer)(v) == nil },
	reflect.Slice:      func(v unsafe.Pointer) bool { return (*(*sliceHeader)(v)).Len == 0 },
	reflect.Bool:       func(v unsafe.Pointer) bool { return !(*(*bool)(v)) },
	reflect.String:     func(v unsafe.Pointer) bool { return *(*string)(v) == "" },
	reflect.Int:        func(v unsafe.Pointer) bool { return *(*int)(v) == 0 },
	reflect.Int8:       func(v unsafe.Pointer) bool { return *(*int8)(v) == 0 },
	reflect.Int16:      func(v unsafe.Pointer) bool { return *(*int16)(v) == 0 },
	reflect.Int32:      func(v unsafe.Pointer) bool { return *(*int32)(v) == 0 },
	reflect.Int64:      func(v unsafe.Pointer) bool { return *(*int64)(v) == 0 },
	reflect.Uint:       func(v unsafe.Pointer) bool { return *(*uint)(v) == 0 },
	reflect.Uint8:      func(v unsafe.Pointer) bool { return *(*uint8)(v) == 0 },
	reflect.Uint16:     func(v unsafe.Pointer) bool { return *(*uint16)(v) == 0 },
	reflect.Uint32:     func(v unsafe.Pointer) bool { return *(*uint32)(v) == 0 },
	reflect.Uint64:     func(v unsafe.Pointer) bool { return *(*uint64)(v) == 0 },
	reflect.Float32:    func(v unsafe.Pointer) bool { return *(*float32)(v) == 0 },
	reflect.Float64:    func(v unsafe.Pointer) bool { return *(*float64)(v) == 0 },
	reflect.Complex64:  func(v unsafe.Pointer) bool { return *(*complex64)(v) == 0 },
	reflect.Complex128: func(v unsafe.Pointer) bool { return *(*complex128)(v) == complex(0, 0) },
}
