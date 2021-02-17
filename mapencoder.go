package jingo

// mapencoder.go manages MapEncoder and its responsibilities.
// MapEnder follows the same principle of structencoder.go in that it generates lightweight
// instructions as part of its compile stage which are executed later during the Marshal.
// The key differences are:
// - a separate instruction for key (kconv) and element (econv)
// - a hashmap iterator is used to iterate over the map
// - accepts `Config` which used to enable/disable sorting keys prior to encoding the map
// - there's fast instruction(s) for sorted & unsorted `map[string]string`

import (
	"bytes"
	"encoding"
	"reflect"
	"sort"
	"sync"
	"unsafe"
)

// MapEncoder stores a set of instructions for building a JSON document from a map at runtime.
type MapEncoder struct {
	instruction func(t unsafe.Pointer, w *Buffer)
	typ         unsafe.Pointer
	ttKey       reflect.Type
	ttElem      reflect.Type
	cfg         Config
}

// NewMapEncoder builds a new MapEncoder.
func NewMapEncoder(t interface{}) *MapEncoder {
	return NewMapEncoderWithConfig(t, DefaultConfig())
}

// Marshal executes the instruction set built up by NewMapEncoder.
func (e MapEncoder) Marshal(s interface{}, w *Buffer) {

	p := unsafe.Pointer(reflect.ValueOf(s).Pointer())
	e.instruction(p, w)
}

var textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

// NewMapEncoderWithConfig builds a new MapEncoder using Config provided.
func NewMapEncoderWithConfig(t interface{}, cfg Config) *MapEncoder {

	e := &MapEncoder{cfg: cfg}

	tt := reflect.TypeOf(t)

	e.typ = (*eface)(unsafe.Pointer(&t)).typ

	e.ttKey = tt.Key()
	e.ttElem = tt.Elem()

	if e.ttKey.Kind() == reflect.Ptr {
		e.ttKey = e.ttKey.Elem()
	}

	if e.ttElem.Kind() == reflect.Ptr {
		e.ttElem = e.ttElem.Elem()
	}

	if tt.Key().Kind() == reflect.String && tt.Elem().Kind() == reflect.String {

		// With optimization:
		// name                                               time/op
		// MapEncoder/Key:_string,_Elem:_string_(sorted)-8    762ns ± 2%
		// MapEncoder/Key:_string,_Elem:_string_(unsorted)-8  306ns ± 0%
		// name                                               alloc/op
		// MapEncoder/Key:_string,_Elem:_string_(sorted)-8    0.00B
		// MapEncoder/Key:_string,_Elem:_string_(unsorted)-8  0.00B
		// name                                               allocs/op
		// MapEncoder/Key:_string,_Elem:_string_(sorted)-8     0.00
		// MapEncoder/Key:_string,_Elem:_string_(unsorted)-8   0.00
		// Without optimization:
		// name                                               time/op
		// MapEncoder/Key:_string,_Elem:_string_(sorted)-8    903ns ± 1%
		// MapEncoder/Key:_string,_Elem:_string_(unsorted)-8  369ns ± 0%
		// name                                               alloc/op
		// MapEncoder/Key:_string,_Elem:_string_(sorted)-8    0.00B
		// MapEncoder/Key:_string,_Elem:_string_(unsorted)-8  0.00B
		// name                                               allocs/op
		// MapEncoder/Key:_string,_Elem:_string_(sorted)-8     0.00
		// MapEncoder/Key:_string,_Elem:_string_(unsorted)-8   0.00

		if e.cfg.SortMapKeys() {
			e.instruction = e.sortStrStrInstr()
			return e
		}

		e.instruction = e.strStrInstr()
		return e
	}

	var econv func(unsafe.Pointer, *Buffer)

	if tt.Elem().Implements(textMarshalerType) {
		if tt.Elem().Kind() == reflect.Ptr {
			econv = e.ptrElemInstr(func(v unsafe.Pointer, w *Buffer) {
				k, _ := reflect.NewAt(e.ttElem, v).Interface().(encoding.TextMarshaler).MarshalText()
				w.WriteByte('"')
				ptrStringToBuf(unsafe.Pointer(&k), w)
				w.WriteByte('"')
			})
		} else {
			econv = func(v unsafe.Pointer, w *Buffer) {
				k, _ := reflect.NewAt(e.ttElem, v).Interface().(encoding.TextMarshaler).MarshalText()
				w.WriteByte('"')
				ptrStringToBuf(unsafe.Pointer(&k), w)
				w.WriteByte('"')
			}
		}

		goto KeyInstr
	}

	switch tt.Elem().Kind() {
	case reflect.Slice:
		enc := NewSliceEncoderWithConfig(reflect.New(tt.Elem()).Elem().Interface(), e.cfg)
		econv = func(v unsafe.Pointer, w *Buffer) {
			var em interface{} = unsafe.Pointer(uintptr(v))
			enc.Marshal(em, w)
		}

	case reflect.Struct:
		enc := NewStructEncoderWithConfig(reflect.New(tt.Elem()).Elem().Interface(), e.cfg)
		econv = func(v unsafe.Pointer, w *Buffer) {
			var em interface{} = unsafe.Pointer(uintptr(v))
			enc.Marshal(em, w)
		}

	case reflect.Map:
		enc := NewMapEncoderWithConfig(reflect.New(tt.Elem()).Elem().Interface(), e.cfg)
		econv = func(v unsafe.Pointer, w *Buffer) {
			var em interface{} = unsafe.Pointer(uintptr(v))
			enc.Marshal(em, w)
		}

	case reflect.Ptr:
		switch tt.Elem().Elem().Kind() {
		case reflect.Slice:
			enc := NewSliceEncoderWithConfig(reflect.New(tt.Elem().Elem()).Elem().Interface(), e.cfg)
			econv = e.ptrElemInstr(func(v unsafe.Pointer, w *Buffer) {
				var em interface{} = unsafe.Pointer(uintptr(v))
				enc.Marshal(em, w)
			})

		case reflect.Struct:
			enc := NewStructEncoderWithConfig(reflect.New(tt.Elem().Elem()).Elem().Interface(), e.cfg)
			econv = e.ptrElemInstr(func(v unsafe.Pointer, w *Buffer) {
				var em interface{} = unsafe.Pointer(uintptr(v))
				enc.Marshal(em, w)
			})

		case reflect.Map:
			enc := NewMapEncoderWithConfig(reflect.New(tt.Elem().Elem()).Elem().Interface(), e.cfg)
			econv = e.ptrElemInstr(func(v unsafe.Pointer, w *Buffer) {
				var em interface{} = unsafe.Pointer(uintptr(v))
				enc.Marshal(em, w)
			})

		case reflect.String:
			econv = e.ptrStrElemInstr()

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
			econv = e.ptrElemInstr(typeconv[tt.Elem().Elem().Kind()])

		default:
			panic("unsupported ptr elem type")
		}

	case reflect.String:
		econv = func(v unsafe.Pointer, w *Buffer) {
			w.WriteByte('"')
			ptrStringToBuf(v, w)
			w.WriteByte('"')
		}

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
		econv = typeconv[tt.Elem().Kind()]

	default:
		panic("unsupported elem type")
	}
KeyInstr:

	var kconv func(unsafe.Pointer, *Buffer)

	switch tt.Key().Kind() {
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
		reflect.Float64,
		reflect.String:
		kconv = typeconv[tt.Key().Kind()]

	default:

		if tt.Key().Implements(textMarshalerType) {
			if tt.Key().Kind() == reflect.Ptr {
				kconv = e.ptrKeyInstr(func(v unsafe.Pointer, w *Buffer) {
					k, _ := reflect.NewAt(e.ttKey, v).Interface().(encoding.TextMarshaler).MarshalText()
					ptrStringToBuf(unsafe.Pointer(&k), w)
				})

				break
			}

			kconv = func(v unsafe.Pointer, w *Buffer) {
				k, _ := reflect.NewAt(e.ttKey, v).Interface().(encoding.TextMarshaler).MarshalText()
				ptrStringToBuf(unsafe.Pointer(&k), w)
			}

			break
		}
		panic("unsupported key type")
	}

	if e.cfg.SortMapKeys() {

		e.instruction = e.sortInstr(kconv, econv)
		return e
	}

	e.instruction = e.instr(kconv, econv)

	return e
}

// ptrStrElemInstr creates an instruction to read from a pointer field we're marshaling
func (e *MapEncoder) ptrStrElemInstr() func(unsafe.Pointer, *Buffer) {
	return func(v unsafe.Pointer, w *Buffer) {

		p := *(*unsafe.Pointer)(v)
		if p == unsafe.Pointer(nil) {
			w.Write(null)
			return
		}

		w.WriteByte('"')
		ptrStringToBuf(p, w)
		w.WriteByte('"')
	}
}

// ptrElemInstr creates an instruction to read from a pointer field we're marshaling
func (e *MapEncoder) ptrElemInstr(conv func(unsafe.Pointer, *Buffer)) func(unsafe.Pointer, *Buffer) {
	return func(v unsafe.Pointer, w *Buffer) {

		p := *(*unsafe.Pointer)(v)
		if p == unsafe.Pointer(nil) {
			w.Write(null)
			return
		}

		conv(p, w)
	}
}

// ptrKeyInstr creates an instruction to read from a pointer field we're marshaling
func (e *MapEncoder) ptrKeyInstr(conv func(unsafe.Pointer, *Buffer)) func(unsafe.Pointer, *Buffer) {
	return func(v unsafe.Pointer, w *Buffer) {

		p := *(*unsafe.Pointer)(v)
		if p == unsafe.Pointer(nil) {
			// write empty quotes for nil keys
			return
		}

		conv(p, w)
	}
}

var emptyObj = []byte("{}")

func (e *MapEncoder) sortStrStrInstr() func(unsafe.Pointer, *Buffer) {

	return func(p unsafe.Pointer, w *Buffer) {

		m := *(*unsafe.Pointer)(p)

		if m == nil {
			w.Write(null)
			return
		}

		mlen := maplen(m)

		if mlen == 0 {
			w.Write(emptyObj)
			return
		}

		it := newhiter(e.typ, m)
		mapSlice := newMapSliceFromPool()

		for ; it.key != nil; mapiternext(it) {

			sl := (*sliceHeader)(it.key)
			mapSlice.kvs = append(mapSlice.kvs, unsafeke{*sl, it.elem})
		}

		hiterPool.Put(it)
		sort.Sort(mapSlice)

		w.Write([]byte(`{"`))

		for i, l := 0, mlen; i < l; i++ {

			if i != 0 {
				w.Write([]byte(`","`))
			}

			ptrStringToBuf(unsafe.Pointer(&mapSlice.kvs[i].k), w)
			w.Write([]byte(`":"`))
			ptrStringToBuf(mapSlice.kvs[i].e, w)
		}

		mapSlice.ReturnToPool()

		w.Write([]byte(`"}`))
	}
}

func (e *MapEncoder) strStrInstr() func(unsafe.Pointer, *Buffer) {

	return func(p unsafe.Pointer, w *Buffer) {

		m := *(*unsafe.Pointer)(p)

		if m == nil {
			w.Write(null)
			return
		}

		if maplen(m) == 0 {
			w.Write(emptyObj)
			return
		}

		w.Write([]byte(`{"`))

		it := newhiter(e.typ, m)

		for i := 0; it.key != nil; mapiternext(it) {

			if i != 0 {
				w.Write([]byte(`","`))
			}

			ptrStringToBuf(it.key, w)
			w.Write([]byte(`":"`))
			ptrStringToBuf(it.elem, w)

			i++
		}

		hiterPool.Put(it)

		w.Write([]byte(`"}`))
	}
}

func (e *MapEncoder) sortInstr(kconv, econv func(unsafe.Pointer, *Buffer)) func(unsafe.Pointer, *Buffer) {

	return func(p unsafe.Pointer, w *Buffer) {

		m := *(*unsafe.Pointer)(p)

		if m == nil {
			w.Write(null)
			return
		}

		mlen := maplen(m)

		if mlen == 0 {
			w.Write(emptyObj)
			return
		}

		var (
			bufStart = len(w.Bytes)
			ptrBuf   = unsafe.Pointer(&w.Bytes)
			sl       = (*sliceHeader)(ptrBuf)
		)

		it := newhiter(e.typ, m)
		mapSlice := newMapSliceFromPool()

		for i := 0; it.key != nil; mapiternext(it) {

			start := len(w.Bytes)
			kconv(it.key, w)

			klen := len(w.Bytes) - start

			mapSlice.kvs = append(mapSlice.kvs,
				unsafeke{
					k: sliceHeader{unsafe.Pointer(uintptr(sl.Data) + uintptr(start)), klen, klen},
					e: it.elem,
				})

			i++
		}

		hiterPool.Put(it)

		bufEnd := len(w.Bytes)

		sort.Sort(mapSlice)

		w.Write([]byte(`{"`))

		for i, l := 0, mlen; i < l; i++ {

			if i != 0 {
				w.Write([]byte(`,"`))
			}

			w.Bytes = append(w.Bytes, *(*[]byte)(unsafe.Pointer(&mapSlice.kvs[i].k))...)
			w.Write([]byte(`":`))
			econv(mapSlice.kvs[i].e, w)
		}
		mapSlice.ReturnToPool()

		w.Bytes = append(w.Bytes[:bufStart], w.Bytes[bufEnd:]...)
		w.WriteByte('}')
	}
}

func (e *MapEncoder) instr(kconv, econv func(unsafe.Pointer, *Buffer)) func(unsafe.Pointer, *Buffer) {

	return func(p unsafe.Pointer, w *Buffer) {

		m := *(*unsafe.Pointer)(p)

		if m == nil {
			w.Write(null)
			return
		}

		if maplen(m) == 0 {
			w.Write(emptyObj)
			return
		}

		w.Write([]byte(`{"`))

		it := newhiter(e.typ, m)

		for i := 0; it.key != nil; mapiternext(it) {

			if i != 0 {
				w.Write([]byte(`,"`))
			}

			kconv(it.key, w)
			w.Write([]byte(`":`))
			econv(it.elem, w)

			i++
		}

		hiterPool.Put(it)
		w.WriteByte('}')
	}
}

// DefaultConfig returns DefaultConfiguration for NewEncoder instructions. By default, map keys are unsorted.
func DefaultConfig() Config {

	c := Config{}
	c.SetSortMapKeys(false)
	return c
}

// Config is a type used to represent configuration options that can be
// applied when formatting json output.
type Config struct {
	mapEncoder uint8
}

const (
	// map encoder
	sortMapKeys uint8 = 1 << iota
)

// SetSortMapKeys specifies whether map keys are sorted before to encoding values to JSON. Setting `SortMapKeys` to off drastically improves performance for MapEncoders.
func (c *Config) SetSortMapKeys(on bool) {
	if on {
		c.mapEncoder |= sortMapKeys
		return
	}

	c.mapEncoder &= ^sortMapKeys
}

// SortMapKeys states whether SortMapKeys is on/off.
func (c Config) SortMapKeys() bool {
	return c.mapEncoder&sortMapKeys != 0
}

type eface struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

// A hash iteration structure.
// Make sure this stays in sync with runtime/map.go.
type hiter struct {
	key  unsafe.Pointer
	elem unsafe.Pointer

	_ unsafe.Pointer //    t           unsafe.Pointer // *MapType
	_ unsafe.Pointer //    h           *hmap
	_ unsafe.Pointer //    buckets     *bmap
	_ unsafe.Pointer //    bptr        *bmap
	_ unsafe.Pointer //    overflow    unsafe.Pointer // *[]*bmap
	_ unsafe.Pointer //    oldoverflow unsafe.Pointer // *[]*bmap
	_ uintptr        //    startBucket uintptr
	_ uint8          //    offset      uint8
	_ bool           //    wrapped     bool
	_ uint8          //    B           uint8
	_ uint8          //    i           uint8
	_ uintptr        //    bucket      uintptr
	_ uintptr        //    checkBucket uintptr
}

var (
	hiterPool sync.Pool
	zeroHiter = &hiter{}
)

func newhiter(t, m unsafe.Pointer) *hiter {
	v := hiterPool.Get()
	if v == nil {
		return mapiterinit(t, m)
	}
	it := v.(*hiter)
	*it = *zeroHiter
	runtimemapiterinit(t, m, unsafe.Pointer(it))
	return it
}

//go:noescape
//go:linkname mapiterinit reflect.mapiterinit
func mapiterinit(unsafe.Pointer, unsafe.Pointer) *hiter

//go:noescape
//go:linkname mapiternext reflect.mapiternext
func mapiternext(*hiter)

//go:noescape
//go:linkname runtimemapiterinit runtime.mapiterinit
func runtimemapiterinit(unsafe.Pointer, unsafe.Pointer, unsafe.Pointer)

//go:noescape
//go:linkname maplen reflect.maplen
func maplen(unsafe.Pointer) int

type mapSlice struct {
	kvs []unsafeke
}

func (ms mapSlice) Len() int {
	return len(ms.kvs)
}

func (ms mapSlice) Swap(i, j int) {
	ms.kvs[i], ms.kvs[j] = ms.kvs[j], ms.kvs[i]
}

func (ms mapSlice) Less(i, j int) bool {
	return bytes.Compare(*(*[]byte)(unsafe.Pointer(&ms.kvs[i].k)), *(*[]byte)(unsafe.Pointer(&ms.kvs[j].k))) < 0
}

func (ms *mapSlice) ReturnToPool() {
	mapSlicePool.Put(ms)
}

func (ms *mapSlice) Reset() {
	ms.kvs = ms.kvs[:0]
}

var mapSlicePool = sync.Pool{New: func() interface{} { return &mapSlice{} }}

func newMapSliceFromPool() *mapSlice {
	newMapSlice := mapSlicePool.Get().(*mapSlice)
	newMapSlice.Reset()

	return newMapSlice
}

type unsafeke struct {
	k sliceHeader
	e unsafe.Pointer
}

