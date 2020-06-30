package jingo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"
)

type all struct {
	PropBool    bool    `json:"propBool"`
	PropInt     int     `json:"propInt"`
	PropInt8    int8    `json:"propInt8"`
	PropInt16   int16   `json:"propInt16"`
	PropInt32   int32   `json:"propInt32"`
	PropInt64   int64   `json:"propInt64"`
	PropUint    uint    `json:"propUint"`
	PropUint8   uint8   `json:"propUint8"`
	PropUint16  uint16  `json:"propUint16"`
	PropUint32  uint32  `json:"propUint32"`
	PropUint64  uint64  `json:"propUint64"`
	PropFloat32 float32 `json:"propFloat32"`
	PropFloat64 float64 `json:"propFloat64,stringer"`
	PropString  string  `json:"propString"`
	PropStruct  struct {
		PropNames []string  `json:"propName"`
		PropPs    []*string `json:"ps"`
	} `json:"propStruct"`
	PropEncode     encode0  `json:"propEncode,encoder"`
	PropEncodeP    *encode0 `json:"propEncodeP,encoder"`
	PropEncodenilP *encode0 `json:"propEncodenilP,encoder"`
	PropEncodeS    encode1  `json:"propEncodeS,encoder"`
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

func ExampleJsonAll() {

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
			PropNames []string  `json:"propName"`
			PropPs    []*string `json:"ps"`
		}{
			PropNames: []string{"a name", "another name", "another"},
			PropPs:    []*string{&s, nil, &s},
		},
		PropEncode:  encode0{'1'},
		PropEncodeP: &encode0{'2'},
		PropEncodeS: encode1{encode0{'3'}, encode0{'4'}},
	}, b)

	fmt.Println(b.String())

	// Output:
	// {"propBool":false,"propInt":1234567878910111212,"propInt8":123,"propInt16":12349,"propInt32":1234567891,"propInt64":1234567878910111213,"propUint":12345678789101112138,"propUint8":255,"propUint16":12345,"propUint32":1234567891,"propUint64":12345678789101112139,"propFloat32":21.232426,"propFloat64":2799999999888.2827,"propString":"thirty two thirty four","propStruct":{"propName":["a name","another name","another"],"ps":["test pointer string",null,"test pointer string"]},"propEncode":1,"propEncodeP":2,"propEncodenilP":null,"propEncodeS":134}
}

func ExampleRaw() {

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

	b.StartTimer()
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

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&ub)
	}
}

type TimeObject struct {
	Time         time.Time    `json:"time"`
	PtrTime      *time.Time   `json:"ptrTime"`
	SliceTime    []time.Time  `json:"sliceTime"`
	PtrSliceTime []*time.Time `json:"ptrSliceTime"`
}

func Test_Time(t *testing.T) {

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

	wantJSON := `{"time":2000-09-17T20:04:26Z,"ptrTime":2001-09-17T20:04:26Z,"sliceTime":["2002-09-17T20:04:26Z"],"ptrSliceTime":["2003-09-17T20:04:26Z"]}`

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

	b.StartTimer()
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

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&ss)
	}
}

func TestStructEncoder_map(t *testing.T) {

	type s0 struct {
		M map[string]string `json:"m"`
	}

	v := s0{map[string]string{"aa": "1", "bb": "2", "cc": "3"}}
	want := []byte(`{"m":{"aa":"1","bb":"2","cc":"3"}}`)

	enc := NewStructEncoder(s0{})

	buf := NewBufferFromPool()
	defer buf.ReturnToPool()

	enc.Marshal(&v, buf)

	if !bytes.Equal(want, buf.Bytes) {
		t.Errorf("\nwant:\n%s\ngot:\n%s\n", want, buf.Bytes)
	}
}

func TestStructEncoder_ptrmap(t *testing.T) {

	type s0 struct {
		M *map[string]string `json:"m"`
	}

	m := map[string]string{"aa": "1", "bb": "2", "cc": "3"}

	v := s0{&m}
	want := []byte(`{"m":{"aa":"1","bb":"2","cc":"3"}}`)

	enc := NewStructEncoder(s0{})

	buf := NewBufferFromPool()
	defer buf.ReturnToPool()

	enc.Marshal(&v, buf)

	if !bytes.Equal(want, buf.Bytes) {
		t.Errorf("\nwant:\n%s\ngot:\n%s\n", want, buf.Bytes)
	}
}

func TestSliceEncoder_map(t *testing.T) {

	v := []map[string]string{
		nil,
		{"KA1": "EA1", "KA2": "EA2", "KA3": "EA3"},
		{"KB1": "EB1", "KB2": "EB2"},
		{},
	}
	want := []byte(`[null,{"KA1":"EA1","KA2":"EA2","KA3":"EA3"},{"KB1":"EB1","KB2":"EB2"},{}]`)

	enc := NewSliceEncoder([]map[string]string{})

	buf := NewBufferFromPool()
	defer buf.ReturnToPool()

	enc.Marshal(&v, buf)

	if !bytes.Equal(want, buf.Bytes) {
		t.Errorf("\nwant:\n%s\ngot:\n%s\n", want, buf.Bytes)
	}
}

func TestSliceEncoder_ptrmap(t *testing.T) {

	m1 := map[string]string{"KA1": "EA1", "KA2": "EA2", "KA3": "EA3"}
	m2 := map[string]string{"KB1": "EB1", "KB2": "EB2"}
	v := []*map[string]string{
		nil,
		&m1,
		&m2,
		{},
	}
	want := []byte(`[null,{"KA1":"EA1","KA2":"EA2","KA3":"EA3"},{"KB1":"EB1","KB2":"EB2"},{}]`)

	enc := NewSliceEncoder([]*map[string]string{})

	buf := NewBufferFromPool()
	defer buf.ReturnToPool()

	enc.Marshal(&v, buf)

	if !bytes.Equal(want, buf.Bytes) {
		t.Errorf("\nwant:\n%s\ngot:\n%s\n", want, buf.Bytes)
	}
}

// var fakeType = SmallPayload{}
// var fake = NewSmallPayload()

// var fakeType = LargePayload{}
// var fake = NewLargePayload()

//
//
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
		PropNames []string  `json:"propName"`
		PropPs    []*string `json:"ps"`
	}{
		PropNames: []string{"a name", "another name", "another"},
		PropPs:    []*string{&s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s, nil, &s},
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
