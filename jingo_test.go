package jingo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type all struct {
	ignoreMe1    string
	PropBool     bool    `json:"propBool"`
	PropInt      int     `json:"propInt"`
	PropInt8     int8    `json:"propInt8"`
	PropInt16    int16   `json:"propInt16"`
	PropInt32    int32   `json:"propInt32"`
	PropInt64    int64   `json:"propInt64"`
	PropUint     uint    `json:"propUint"`
	PropUint8    uint8   `json:"propUint8"`
	PropUint16   uint16  `json:"propUint16"`
	PropUint32   uint32  `json:"propUint32"`
	PropUint64   uint64  `json:"propUint64"`
	PropFloat32  float32 `json:"propFloat32"`
	PropFloat64  float64 `json:"propFloat64,stringer"`
	PropString   string  `json:"propString"`
	ignoreStruct struct {
		ignoreMeStruct1 string
		ignoreMeStruct2 string
		ignoreMeStruct3 string
	}
	PropStruct struct {
		PropNames        []string  `json:"propName"`
		PropPs           []*string `json:"ps"`
		PropNamesEscaped []string  `json:"propNameEscaped,escape"`
	} `json:"propStruct"`
	PropEncode         encode0       `json:"propEncode,encoder"`
	PropEncodeP        *encode0      `json:"propEncodeP,encoder"`
	PropEncodenilP     *encode0      `json:"propEncodenilP,encoder"`
	PropEncodeS        encode1       `json:"propEncodeS,encoder"`
	PropJSONMarshaler  jsonMarshaler `json:"propJSONMarshaler,encoder"`
	ignoreMe2          string
	PropJSONMarshalerP *jsonMarshaler `json:"propJSONMarshalerP,encoder"`
	ignoreMe3          string
}

type allSubOmit struct {
	PropBool   bool   `json:"propBool,omitempty"`
	PropInt    int    `json:"propInt,omitempty"`
	PropString string `json:"propString,omitempty"`
}

type allOmit struct {
	ignoreMe1    string
	PropBool     bool       `json:"propBool,omitempty"`
	PropInt      int        `json:"propInt,omitempty"`
	PropInt8     int8       `json:"propInt8,omitempty"`
	PropInt16    int16      `json:"propInt16,omitempty"`
	PropInt32    int32      `json:"propInt32,omitempty"`
	PropInt64    int64      `json:"propInt64,omitempty"`
	PropUint     uint       `json:"propUint,omitempty"`
	PropUint8    uint8      `json:"propUint8,omitempty"`
	PropUint16   uint16     `json:"propUint16,omitempty"`
	PropUint32   uint32     `json:"propUint32,omitempty"`
	PropUint64   uint64     `json:"propUint64,omitempty"`
	PropFloat32  float32    `json:"propFloat32,omitempty"`
	PropFloat64  float64    `json:"propFloat64,stringer,omitempty"`
	PropString   string     `json:"propString,omitempty"`
	PropStruct   allSubOmit `json:"propStruct,omitempty"`
	ignoreStruct struct {
		ignoreMeStruct1 string
		ignoreMeStruct2 string
		ignoreMeStruct3 string
	}
	PropPointerStruct  *allSubOmit       `json:"propPointerStruct,omitempty"`
	PropSlice          []string          `json:"propName,omitempty"`
	PropPointerSlice   []*string         `json:"ps,omitempty"`
	PropSliceEscaped   []string          `json:"propNameEscaped,escape,omitempty"`
	PropEncode         encodeOmit0       `json:"propEncode,encoder,omitempty"`
	PropEncodeP        *encodeOmit0      `json:"propEncodeP,encoder,omitempty"`
	PropEncodenilP     *encodeOmit0      `json:"propEncodenilP,encoder,omitempty"`
	PropEncodeS        encodeOmit1       `json:"propEncodeS,encoder,omitempty"`
	PropJSONMarshaler  jsonMarshalerOmit `json:"propJSONMarshaler,encoder,omitempty"`
	ignoreMe3          string
	PropJSONMarshalerP *jsonMarshalerOmit `json:"propJSONMarshalerP,encoder,omitempty"`
	ignoreMe2          string
}

type encode0 struct {
	val byte
}

func (e *encode0) JSONEncode(w *Buffer) {
	w.WriteByte(e.val)
}

type encode1 []encode0

func (e *encode1) JSONEncode(w *Buffer) {

	w.WriteByte('1')

	for _, v := range *e {
		w.WriteByte(v.val)
	}
}

type encodeOmit0 struct {
	val byte
}

func (e *encodeOmit0) JSONEncode(w *Buffer) {
	if e.val == 0 {
		w.Write([]byte{'{', '}'})
		return
	}

	w.WriteByte(e.val)
}

type encodeOmit1 []encodeOmit0

func (e *encodeOmit1) JSONEncode(w *Buffer) {
	if len(*e) == 0 {
		w.Write([]byte{'{', '}'})
		return
	}

	for _, v := range *e {
		w.WriteByte(v.val)
	}
}

type jsonMarshaler struct {
	val []byte
}

func (j *jsonMarshaler) EncodeJSON(w io.Writer) {
	w.Write(j.val)
}

type jsonMarshalerOmit struct {
	val []byte
}

func (j *jsonMarshalerOmit) EncodeJSON(w io.Writer) {
	if len(j.val) == 0 {
		w.Write([]byte{'{', '}'})
		return
	}
	w.Write(j.val)
}

func Example() {

	enc := NewStructEncoder(all{})
	b := NewBufferFromPool()

	s := "test pointer string"
	enc.Marshal(&all{
		PropBool:    false,
		PropInt:     1234567878910111212,
		PropInt8:    123,
		PropInt16:   12349,
		PropInt32:   1234567891,
		PropInt64:   1234567878910111213,
		PropUint:    12345678789101112138,
		PropUint8:   255,
		PropUint16:  12345,
		PropUint32:  1234567891,
		PropUint64:  12345678789101112139,
		PropFloat32: 21.232426,
		PropFloat64: 2799999999888.28293031999999,
		PropString:  "thirty two thirty four",
		PropStruct: struct {
			PropNames        []string  `json:"propName"`
			PropPs           []*string `json:"ps"`
			PropNamesEscaped []string  `json:"propNameEscaped,escape"`
		}{
			PropNames:        []string{"a name", "another name", "another"},
			PropPs:           []*string{&s, nil, &s},
			PropNamesEscaped: []string{"one\\two\\,three\"", "\"four\\five\\,six\""},
		},
		PropEncode:         encode0{'1'},
		PropEncodeP:        &encode0{'2'},
		PropEncodeS:        encode1{encode0{'3'}, encode0{'4'}},
		PropJSONMarshaler:  jsonMarshaler{[]byte("1")},
		PropJSONMarshalerP: &jsonMarshaler{[]byte("2")},

		ignoreMe1: "1",
		ignoreMe2: "2",
		ignoreMe3: "3",

		ignoreStruct: struct {
			ignoreMeStruct1 string
			ignoreMeStruct2 string
			ignoreMeStruct3 string
		}{ignoreMeStruct1: "", ignoreMeStruct2: "", ignoreMeStruct3: ""},
	}, b)

	fmt.Println(b.String())

	// Output:
	// {"propBool":false,"propInt":1234567878910111212,"propInt8":123,"propInt16":12349,"propInt32":1234567891,"propInt64":1234567878910111213,"propUint":12345678789101112138,"propUint8":255,"propUint16":12345,"propUint32":1234567891,"propUint64":12345678789101112139,"propFloat32":21.232426,"propFloat64":2799999999888.2827,"propString":"thirty two thirty four","propStruct":{"propName":["a name","another name","another"],"ps":["test pointer string",null,"test pointer string"],"propNameEscaped":["one\\two\\,three\"","\"four\\five\\,six\""]},"propEncode":1,"propEncodeP":2,"propEncodenilP":null,"propEncodeS":134,"propJSONMarshaler":1,"propJSONMarshalerP":2}
}

func Example_testStruct2() {

	type testStruct2 struct {
		Raw  []byte `json:"raw,raw"`
		Raw2 []byte `json:"c,raw"`
		Raw3 int    `json:"b,raw"`
	}

	var enc = NewStructEncoder(testStruct2{})

	b := NewBufferFromPool()
	v := testStruct2{
		Raw:  []byte(`{"mapKey1":1,"mapKey2":2}`),
		Raw3: 1,
	}

	enc.Marshal(&v, b)
	fmt.Println(b.String())

	// Output:
	// {"raw":{"mapKey1":1,"mapKey2":2},"c":null,"b":null}
}

func Test_NilStruct(t *testing.T) {
	type testStruct1 struct {
		StrVal string `json:"str1"`
		IntVal int    `json:"int1"`
	}
	type testStruct0 struct {
		StructPtr *testStruct1 `json:"structPtr"`
	}

	wantJSON := "{\"structPtr\":null}"

	var enc = NewStructEncoder(testStruct0{})

	buf := NewBufferFromPool()
	v := testStruct0{}
	enc.Marshal(&v, buf)

	resultJSON := buf.String()
	if resultJSON != wantJSON {
		t.Errorf("Test_NilStruct Failed: want JSON: " + wantJSON + " got JSON:" + resultJSON)
	}
}

type StructWithEscapes struct {
	String      string   `json:"str,escape"`
	StringArray []string `json:"str-array,escape"`
}

func Test_StructWithEscapes(t *testing.T) {
	es := StructWithEscapes{
		String: `one\two\,three,
		four"`, //N.B. includes 2 indentation tabs
		StringArray: []string{`one\two`, `three\,four,
		five`, //N.B. includes 2 indentation tabs
		},
	}

	wantJSON := `{"str":"one\\two\\,three,\n\t\tfour\"","str-array":["one\\two","three\\,four,\n\t\tfive"]}`

	var enc = NewStructEncoder(StructWithEscapes{})
	buf := NewBufferFromPool()
	enc.Marshal(&es, buf)
	resultJSON := buf.String()

	// Ensure JSON is valid.
	if !json.Valid([]byte(resultJSON)) {
		t.Errorf("Not valid JSON:" + resultJSON)
	}

	// Compare result
	if resultJSON != wantJSON {
		t.Errorf("Test_StructWithEscapes Failed: want JSON:" + wantJSON + " got JSON:" + resultJSON)
	}

	andBackAgain := StructWithEscapes{}
	json.Unmarshal([]byte(resultJSON), &andBackAgain)

	if !reflect.DeepEqual(es, andBackAgain) {
		t.Errorf("Test_StructWithEscapes Failed: want: %+v got: %+v", es, andBackAgain)
	}
}

func Test_StructWithRecursion(t *testing.T) {

	type structWithRecursion struct {
		Name  string               `json:"name"`
		Child *structWithRecursion `json:"child"`
	}

	v := structWithRecursion{
		Name: "A",
		Child: &structWithRecursion{
			Name: "B",
			Child: &structWithRecursion{
				Name:  "C",
				Child: nil,
			},
		},
	}

	wantJSON := `{"name":"A","child":{"name":"B","child":{"name":"C","child":null}}}`

	var enc = NewStructEncoder(structWithRecursion{})
	buf := NewBufferFromPool()
	enc.Marshal(&v, buf)

	resultJSON := buf.String()
	if resultJSON != wantJSON {
		t.Errorf("Test_StructWithRecursion Failed: want JSON:" + wantJSON + " got JSON:" + resultJSON)
	}
}

type UnicodeObject struct {
	Chinese string `json:"chinese"`
	Emoji   string `json:"emoji"`
	Russian string `json:"russian"`
}

func Test_Unicode(t *testing.T) {
	ub := UnicodeObject{
		Chinese: "你好，世界",
		Emoji:   "👋🌍😄😂👋💊🐂🍺",
		Russian: "ру́сский язы́к",
	}

	wantJSON := "{\"chinese\":\"你好，世界\",\"emoji\":\"👋🌍😄😂👋💊🐂🍺\",\"russian\":\"ру́сский язы́к\"}"

	var enc = NewStructEncoder(UnicodeObject{})
	buf := NewBufferFromPool()
	enc.Marshal(&ub, buf)
	resultJSON := buf.String()
	if resultJSON != wantJSON {
		t.Errorf("Test_UnicodeEncode Failed: want JSON:" + wantJSON + " got JSON:" + resultJSON)
	}

}

func BenchmarkUnicode(b *testing.B) {
	ub := UnicodeObject{
		Chinese: "你好，世界",
		Emoji:   "👋🌍😄😂💊🐂🍺",
		Russian: "ру́сский язы́к",
	}

	var enc = NewStructEncoder(UnicodeObject{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewBufferFromPool()
		enc.Marshal(&ub, buf)
		buf.ReturnToPool()
	}
}

func BenchmarkUnicodeStdLib(b *testing.B) {
	ub := UnicodeObject{
		Chinese: "你好，世界",
		Emoji:   "👋🌍😄😂💊🐂🍺",
		Russian: "ру́сский язы́к",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&ub)
	}
}

type UnicodeObjectLarge struct {
	Chinese string `json:"chinese"`
	Emoji   string `json:"emoji"`
	Russian string `json:"russian"`
	Test    string `json:"test"`
	Test1   string `json:"test1"`
	Test2   string `json:"test2"`
	Test3   string `json:"test3"`
}

func Test_UnicodeLarge(t *testing.T) {
	ub := UnicodeObjectLarge{
		Chinese: "你好，世界",
		Emoji:   "👋🌍😄😂👋💊🐂🍺",
		Russian: "ру́сский язы́к",
		Test:    "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
		Test1:   "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
		Test2:   "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
		Test3:   "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
	}
	// return

	wantJSON := `{"chinese":"你好，世界","emoji":"👋🌍😄😂👋💊🐂🍺","russian":"ру́сский язы́к","test":"ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ","test1":"ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ","test2":"ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ","test3":"ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl "}`

	var enc = NewStructEncoder(UnicodeObjectLarge{})
	buf := NewBufferFromPool()
	enc.Marshal(&ub, buf)
	resultJSON := buf.String()

	if !json.Valid(buf.Bytes) {
		panic("not valid json")
	}

	if resultJSON != wantJSON {
		t.Errorf("Test_UnicodeEncode Failed: want JSON:" + wantJSON + " got JSON:" + resultJSON)
	}

}

func BenchmarkUnicodeLarge(b *testing.B) {
	ub := UnicodeObjectLarge{
		Chinese: "你好，世界",
		Emoji:   "👋🌍😄😂👋💊🐂🍺",
		Russian: "ру́сский язы́к",
		Test:    "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
		Test1:   "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
		Test2:   "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
		Test3:   "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
	}

	var enc = NewStructEncoder(UnicodeObjectLarge{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewBufferFromPool()
		enc.Marshal(&ub, buf)
		buf.ReturnToPool()
	}
}

func BenchmarkUnicodeLargeStdLib(b *testing.B) {
	ub := UnicodeObjectLarge{
		Chinese: "你好，世界",
		Emoji:   "👋🌍😄😂👋💊🐂🍺",
		Russian: "ру́сский язы́к",
		Test:    "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
		Test1:   "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
		Test2:   "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
		Test3:   "ascdjkl ascdhjklacdshlacdshjkl acdshjcdhjkl acdshjl kacdshjkl acdshjkacdshjklacdhjskl hjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjk lacdshjk acdshjkl acdshjkl hjkl acdshjkl acdshjkl acdshjkl cdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl acdshjkl ",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&ub)
	}
}

type TimeObject struct {
	Time             time.Time    `json:"time"`
	PtrTime          *time.Time   `json:"ptrTime"`
	NullPtrTime      *time.Time   `json:"nullPtrTime"`
	SliceTime        []time.Time  `json:"sliceTime"`
	PtrSliceTime     []*time.Time `json:"ptrSliceTime"`
	NullPtrSliceTime []*time.Time `json:"nullPtrSliceTime"`
}

func Test_Time(t *testing.T) {

	d0 := time.Date(2000, 9, 17, 20, 4, 26, 0, time.UTC)
	d1 := time.Date(2001, 9, 17, 20, 4, 26, 0, time.UTC)
	d2 := time.Date(2002, 9, 17, 20, 4, 26, 0, time.UTC)
	d3 := time.Date(2003, 9, 17, 20, 4, 26, 0, time.UTC)

	to := TimeObject{
		Time:             d0,
		PtrTime:          &d1,
		NullPtrTime:      nil,
		SliceTime:        []time.Time{d2},
		PtrSliceTime:     []*time.Time{&d3},
		NullPtrSliceTime: []*time.Time{nil},
	}

	wantJSON := `{"time":"2000-09-17T20:04:26Z","ptrTime":"2001-09-17T20:04:26Z","nullPtrTime":null,"sliceTime":["2002-09-17T20:04:26Z"],"ptrSliceTime":["2003-09-17T20:04:26Z"],"nullPtrSliceTime":[null]}`

	var enc = NewStructEncoder(TimeObject{})

	buf := NewBufferFromPool()
	defer buf.ReturnToPool()
	enc.Marshal(&to, buf)
	resultJSON := buf.String()
	if resultJSON != wantJSON {
		t.Errorf("Test_Time Failed: want JSON:" + wantJSON + " got JSON:" + resultJSON)
	}
}

func BenchmarkTime(b *testing.B) {
	b.ReportAllocs()

	d0 := time.Date(2000, 9, 17, 20, 4, 26, 0, time.UTC)
	d1 := time.Date(2001, 9, 17, 20, 4, 26, 0, time.UTC)
	d2 := time.Date(2002, 9, 17, 20, 4, 26, 0, time.UTC)
	d3 := time.Date(2003, 9, 17, 20, 4, 26, 0, time.UTC)

	to := TimeObject{
		Time:         d0,
		PtrTime:      &d1,
		SliceTime:    []time.Time{d2},
		PtrSliceTime: []*time.Time{&d3},
	}

	var enc = NewStructEncoder(TimeObject{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewBufferFromPool()
		enc.Marshal(&to, buf)
		buf.ReturnToPool()
	}
}

func BenchmarkTimeStdLib(b *testing.B) {
	b.ReportAllocs()

	d0 := time.Date(2000, 9, 17, 20, 4, 26, 0, time.UTC)
	d1 := time.Date(2001, 9, 17, 20, 4, 26, 0, time.UTC)
	d2 := time.Date(2002, 9, 17, 20, 4, 26, 0, time.UTC)
	d3 := time.Date(2003, 9, 17, 20, 4, 26, 0, time.UTC)

	to := TimeObject{
		Time:         d0,
		PtrTime:      &d1,
		SliceTime:    []time.Time{d2},
		PtrSliceTime: []*time.Time{&d3},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&to)
	}
}

func TestStructOmitempty(t *testing.T) {

	enc5 := NewStructEncoder(allOmit{})

	type marshaler interface {
		Marshal(s interface{}, w *Buffer)
	}

	tests := []struct {
		name string
		enc  marshaler
		v    interface{}
		want []byte
		json bool
	}{
		{
			"No Omit",
			enc5,
			&allOmit{
				PropBool:    true,
				PropInt:     1234567878910111212,
				PropInt8:    123,
				PropInt16:   12349,
				PropInt32:   1234567891,
				PropInt64:   1234567878910111213,
				PropUint:    12345678789101112138,
				PropUint8:   255,
				PropUint16:  12345,
				PropUint32:  1234567891,
				PropUint64:  12345678789101112139,
				PropFloat32: 21.232426,
				PropFloat64: 2799999999888.28293031999999,
				PropString:  "thirty two thirty four",
				PropStruct: allSubOmit{
					PropBool:   true,
					PropInt:    1234567878910111212,
					PropString: "thirty two thirty four",
				},
				PropPointerStruct: &allSubOmit{
					PropBool:   true,
					PropInt:    1234567878910111212,
					PropString: "thirty two thirty four",
				},
				PropSlice:          []string{"a name", "another name", "another"},
				PropPointerSlice:   []*string{&s, nil, &s},
				PropSliceEscaped:   []string{"one\\two\\,three\"", "\"four\\five\\,six\""},
				PropEncode:         encodeOmit0{'1'},
				PropEncodeP:        &encodeOmit0{'2'},
				PropEncodeS:        encodeOmit1{encodeOmit0{'3'}, encodeOmit0{'4'}},
				PropJSONMarshaler:  jsonMarshalerOmit{[]byte("1")},
				PropJSONMarshalerP: &jsonMarshalerOmit{[]byte("2")},

				ignoreMe1: "1",
				ignoreMe2: "2",
				ignoreMe3: "3",

				ignoreStruct: struct {
					ignoreMeStruct1 string
					ignoreMeStruct2 string
					ignoreMeStruct3 string
				}{ignoreMeStruct1: "", ignoreMeStruct2: "", ignoreMeStruct3: ""},
			},
			[]byte(`{"propBool":true,"propInt":1234567878910111212,"propInt8":123,"propInt16":12349,"propInt32":1234567891,"propInt64":1234567878910111213,"propUint":12345678789101112138,"propUint8":255,"propUint16":12345,"propUint32":1234567891,"propUint64":12345678789101112139,"propFloat32":21.232426,"propFloat64":2799999999888.2827,"propString":"thirty two thirty four","propStruct":{"propBool":true,"propInt":1234567878910111212,"propString":"thirty two thirty four"},"propPointerStruct":{"propBool":true,"propInt":1234567878910111212,"propString":"thirty two thirty four"},"propName":["a name","another name","another"],"ps":["test pointer string b",null,"test pointer string b"],"propNameEscaped":["one\\two\\,three\"","\"four\\five\\,six\""],"propEncode":1,"propEncodeP":2,"propEncodeS":34,"propJSONMarshaler":1,"propJSONMarshalerP":2}`),
			false,
		},
		{
			"Omit",
			enc5,
			&allOmit{},
			[]byte(`{"propStruct":{},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Empty Omit",
			enc5,
			&allOmit{
				PropBool:           false,
				PropInt:            0,
				PropInt8:           0,
				PropInt16:          0,
				PropInt32:          0,
				PropInt64:          0,
				PropUint:           0,
				PropUint8:          0,
				PropUint16:         0,
				PropUint32:         0,
				PropUint64:         0,
				PropFloat32:        0,
				PropFloat64:        0,
				PropString:         "",
				PropStruct:         allSubOmit{},
				PropPointerStruct:  &allSubOmit{},
				PropSlice:          []string{},
				PropPointerSlice:   []*string{},
				PropSliceEscaped:   []string{},
				PropEncode:         encodeOmit0{},
				PropEncodeP:        &encodeOmit0{},
				PropEncodeS:        encodeOmit1{},
				PropJSONMarshaler:  jsonMarshalerOmit{},
				PropJSONMarshalerP: &jsonMarshalerOmit{},
			},
			[]byte(`{"propStruct":{},"propPointerStruct":{},"propEncode":{},"propEncodeP":{},"propJSONMarshaler":{},"propJSONMarshalerP":{}}`),
			true,
		},
		{
			"Omit Except String",
			enc5,
			&allOmit{
				PropString: "noomit",
			},
			[]byte(`{"propString":"noomit","propStruct":{},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Int",
			enc5,
			&allOmit{
				PropInt: 365,
			},
			[]byte(`{"propInt":365,"propStruct":{},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Struct",
			enc5,
			&allOmit{
				PropStruct: allSubOmit{
					PropBool:   true,
					PropInt:    1234567878910111212,
					PropString: "thirty two thirty four",
				},
			},
			[]byte(`{"propStruct":{"propBool":true,"propInt":1234567878910111212,"propString":"thirty two thirty four"},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Struct With Bool",
			enc5,
			&allOmit{
				PropStruct: allSubOmit{
					PropBool: true,
				},
			},
			[]byte(`{"propStruct":{"propBool":true},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Struct With String",
			enc5,
			&allOmit{
				PropStruct: allSubOmit{
					PropString: "thirty two thirty four",
				},
			},
			[]byte(`{"propStruct":{"propString":"thirty two thirty four"},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Struct With Int",
			enc5,
			&allOmit{
				PropStruct: allSubOmit{
					PropInt: 1234567878910111212,
				},
			},
			[]byte(`{"propStruct":{"propInt":1234567878910111212},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Slice",
			enc5,
			&allOmit{
				PropSlice: []string{"a name", "another name", "another"},
			},
			[]byte(`{"propStruct":{},"propName":["a name","another name","another"],"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Pointer Slice",
			enc5,
			allOmit{
				PropPointerSlice: []*string{&s, nil, &s},
			},
			[]byte(`{"propStruct":{},"ps":["test pointer string b",null,"test pointer string b"],"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Pointer Struct",
			enc5,
			&allOmit{
				PropPointerStruct: &allSubOmit{
					PropBool:   true,
					PropInt:    1234567878910111212,
					PropString: "thirty two thirty four",
				},
			},
			[]byte(`{"propStruct":{},"propPointerStruct":{"propBool":true,"propInt":1234567878910111212,"propString":"thirty two thirty four"},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Pointer Struct With String",
			enc5,
			&allOmit{
				PropPointerStruct: &allSubOmit{
					PropString: "thirty two thirty four",
				},
			},
			[]byte(`{"propStruct":{},"propPointerStruct":{"propString":"thirty two thirty four"},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
		{
			"Omit Except Pointer Struct With Int",
			enc5,
			&allOmit{
				PropPointerStruct: &allSubOmit{
					PropInt: 1234567878910111212,
				},
			},
			[]byte(`{"propStruct":{},"propPointerStruct":{"propInt":1234567878910111212},"propEncode":{},"propJSONMarshaler":{}}`),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			tt.enc.Marshal(tt.v, buf)

			j, _ := json.Marshal(tt.v)

			if !bytes.Equal(tt.want, buf.Bytes) || (tt.json && !bytes.Equal(j, buf.Bytes)) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\njson:\n%s\n", tt.want, buf.Bytes, j)
			}
		})
	}
}

type omitBenchmark struct {
	PropBool           bool           `json:"propBool,omitempty"`
	PropInt            int            `json:"propInt,omitempty"`
	PropInt8           int8           `json:"propInt8,omitempty"`
	PropInt16          int16          `json:"propInt16,omitempty"`
	PropInt32          int32          `json:"propInt32,omitempty"`
	PropInt64          int64          `json:"propInt64,omitempty"`
	PropUint           uint           `json:"propUint,omitempty"`
	PropUint8          uint8          `json:"propUint8,omitempty"`
	PropUint16         uint16         `json:"propUint16,omitempty"`
	PropUint32         uint32         `json:"propUint32,omitempty"`
	PropUint64         uint64         `json:"propUint64,omitempty"`
	PropFloat32        float32        `json:"propFloat32,omitempty"`
	PropFloat64        float64        `json:"propFloat64,stringer,omitempty"`
	PropString         string         `json:"propString,omitempty"`
	PropPointerStruct  *allSubOmit    `json:"propPointerStruct,omitempty"`
	PropSlice          []string       `json:"propName,omitempty"`
	PropPointerSlice   []*string      `json:"ps,omitempty"`
	PropSliceEscaped   []string       `json:"propNameEscaped,escape,omitempty"`
	PropEncodeP        *encode0       `json:"propEncodeP,encoder,omitempty"`
	PropEncodenilP     *encode0       `json:"propEncodenilP,encoder,omitempty"`
	PropJSONMarshalerP *jsonMarshaler `json:"propJSONMarshalerP,encoder,omitempty"`
}

var omitBench = omitBenchmark{}

func BenchmarkOmitEmpty(b *testing.B) {
	var enc = NewStructEncoder(omitBenchmark{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewBufferFromPool()
		enc.Marshal(&omitBench, buf)
		buf.ReturnToPool()
	}
}

func BenchmarkOmitEmptyStdLib(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&omitBench)
	}
}

var omitBenchWithData = omitBenchmark{
	PropBool:    true,
	PropInt:     1234567878910111212,
	PropInt8:    123,
	PropInt16:   12349,
	PropInt32:   1234567891,
	PropInt64:   1234567878910111213,
	PropUint:    12345678789101112138,
	PropUint8:   255,
	PropUint16:  12345,
	PropUint32:  1234567891,
	PropUint64:  12345678789101112139,
	PropFloat32: 21.232426,
	PropFloat64: 2799999999888.28293031999999,
	PropString:  "thirty two thirty four",
	PropPointerStruct: &allSubOmit{
		PropBool:   true,
		PropInt:    1234567878910111212,
		PropString: "thirty two thirty four",
	},
	PropSlice:          []string{"a name", "another name", "another"},
	PropPointerSlice:   []*string{&s, nil, &s},
	PropSliceEscaped:   []string{"one\\two\\,three\"", "\"four\\five\\,six\""},
	PropEncodeP:        &encode0{'2'},
	PropJSONMarshalerP: &jsonMarshaler{[]byte("2")},
}

func BenchmarkOmitEmptyNonEmpty(b *testing.B) {
	var enc = NewStructEncoder(omitBenchmark{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewBufferFromPool()
		enc.Marshal(&omitBenchWithData, buf)
		buf.ReturnToPool()
	}
}

func BenchmarkOmitEmptyNonEmptyStdLib(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&omitBenchWithData)
	}
}

type OmitSmallPayload struct {
	St   int    `json:"st,omitempty"`
	Sid  int    `json:"sid,omitempty"`
	Tt   string `json:"tt,omitempty"`
	Gr   int    `json:"gr,omitempty"`
	UUID string `json:"uuid,omitempty"`
	IP   string `json:"ip,omitempty"`
	Ua   string `json:"ua,omitempty"`
	Tz   int    `json:"tz,omitempty"`
	V    int    `json:"v,omitempty"`
}

var omitSmallBench = OmitSmallPayload{}

func BenchmarkOmitEmptySmall(b *testing.B) {
	var enc = NewStructEncoder(OmitSmallPayload{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewBufferFromPool()
		enc.Marshal(&omitSmallBench, buf)
		buf.ReturnToPool()
	}
}

func BenchmarkOmitEmptySmallStdLib(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&omitSmallBench)
	}
}

func TestSliceEncoder(t *testing.T) {

	enc := NewSliceEncoder([]string{})

	type marshaler interface {
		Marshal(s interface{}, w *Buffer)
	}

	tests := []struct {
		name string
		enc  marshaler
		v    interface{}
		want []byte
	}{
		{
			"SliceEncoder String - Empty",
			enc,
			&[]string{},
			[]byte("[]"),
		},
		{
			"SliceEncoder String - Single",
			enc,
			&[]string{"0"},
			[]byte(`["0"]`),
		},
		{
			"SliceEncoder String - Many",
			enc,
			&[]string{"0", "1", "2"},
			[]byte(`["0","1","2"]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			tt.enc.Marshal(tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s", tt.want, buf.Bytes)
			}

		})
	}
}

func BenchmarkSlice(b *testing.B) {

	ss := []string{
		"a name",
		"another name",
		"another",
		"and one more",
		"last one, promise",
	}

	var enc = NewSliceEncoder([]string{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewBufferFromPool()
		enc.Marshal(&ss, buf)
		buf.ReturnToPool()
	}
}

func BenchmarkSliceEscape(b *testing.B) {

	ss := []string{
		"a name",
		"another name",
		"another",
		"and one more",
		"last one, promise",
	}

	var enc = NewSliceEncoder([]EscapeString{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewBufferFromPool()
		enc.Marshal(&ss, buf)
		buf.ReturnToPool()
	}
}

func BenchmarkSliceStdLib(b *testing.B) {
	ss := []string{
		"a name",
		"another name",
		"another",
		"and one more",
		"last one, promise",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&ss)
	}
}

// var fakeType = SmallPayload{}
// var fake = NewSmallPayload()

// var fakeType = LargePayload{}
// var fake = NewLargePayload()

var s = "test pointer string b"
var fakeType = all{}
var fake = &all{
	PropBool:    false,
	PropInt:     1234567878910111212,
	PropInt8:    123,
	PropInt16:   12349,
	PropInt32:   1234567891,
	PropInt64:   1234567878910111213,
	PropUint:    12345678789101112138,
	PropUint8:   255,
	PropUint16:  12345,
	PropUint32:  1234567891,
	PropUint64:  12345678789101112139,
	PropFloat32: 21.232426,
	PropFloat64: 2799999999888.28293031999999,
	PropString:  "thirty two thirty four",
	PropStruct: struct {
		PropNames        []string  `json:"propName"`
		PropPs           []*string `json:"ps"`
		PropNamesEscaped []string  `json:"propNameEscaped,escape"`
	}{
		PropNames:        []string{"a name", "another name", "another"},
		PropPs:           []*string{&s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s},
		PropNamesEscaped: []string{"one\\two\\,three\"", "\"four\\five\\,six\""},
	},
}

func BenchmarkJson(b *testing.B) {

	e := NewStructEncoder(fakeType)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewBufferFromPool()
		e.Marshal(fake, buf)
		buf.ReturnToPool()
	}
}

func BenchmarkStdJson(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		by, _ := json.Marshal(fake)
		_ = by
	}
}

//
//
//

type SmallPayload struct {
	St   int    `json:"st"`
	Sid  int    `json:"sid"`
	Tt   string `json:"tt"`
	Gr   int    `json:"gr"`
	UUID string `json:"uuid"`
	IP   string `json:"ip"`
	Ua   string `json:"ua"`
	Tz   int    `json:"tz"`
	V    int    `json:"v"`
}

func NewSmallPayload() *SmallPayload {
	s := &SmallPayload{
		St:   1,
		Sid:  2,
		Tt:   "TestString",
		Gr:   4,
		UUID: "8f9a65eb-4807-4d57-b6e0-bda5d62f1429",
		IP:   "127.0.0.1",
		Ua:   "Mozilla",
		Tz:   8,
		V:    6,
	}
	return s
}

var smallPayload = NewSmallPayload()

func BenchmarkSmallPayload(b *testing.B) {

	e := NewStructEncoder(SmallPayload{})

	buf := NewBufferFromPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Marshal(smallPayload, buf)
		buf.Reset()
	}
}

func BenchmarkSmallPayloadStdLib(b *testing.B) {

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(smallPayload)
	}
}

var largePayload = NewLargePayload()

func BenchmarkLargePayload(b *testing.B) {

	e := NewStructEncoder(LargePayload{})
	buf := NewBufferFromPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Marshal(largePayload, buf)
		buf.Reset()
	}
}

func BenchmarkLargePayloadStdLib(b *testing.B) {

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(largePayload)
	}
}

//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//

type DSUser struct {
	Username string `json:"username"`
}

type DSTopic struct {
	ID   int    `json:"ID"`
	Slug string `json:"slug"`
}

type DSTopics []*DSTopic

type DSTopicsList struct {
	Topics        DSTopics `json:"topics"`
	MoreTopicsURL string   `json:"more_topics_URL"`
}

type DSUsers []*DSUser

type LargePayload struct {
	Users  DSUsers       `json:"users"`
	Topics *DSTopicsList `json:"topics"`
}

func NewLargePayload() *LargePayload {
	dsUsers := DSUsers{}
	dsTopics := DSTopics{}
	for i := 0; i < 100; i++ {
		str := "test" + strconv.Itoa(i)
		dsUsers = append(
			dsUsers,
			&DSUser{
				Username: str,
			},
		)
		dsTopics = append(
			dsTopics,
			&DSTopic{
				ID:   i,
				Slug: str,
			},
		)
	}
	return &LargePayload{
		Users: dsUsers,
		Topics: &DSTopicsList{
			Topics:        dsTopics,
			MoreTopicsURL: "http://test.com",
		},
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
