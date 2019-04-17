package jingo

// ptrconvert.go declares a number of primitive form -> buffer conversion
// functions based on an unsafe.Pointer input. We're using the implementation
// from the standard library here, which don't perform badly but are a
// candidate for a more high performance implementation to be introduced.

import (
	"reflect"
	"strconv"
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
	}
	b.Write(bfalse)
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

func ptrStringToBuf(v unsafe.Pointer, b *Buffer) {
	b.Write(*(*[]byte)(v))
}
