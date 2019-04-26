package jingo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
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
	}, b)

	fmt.Println(b.String())

	// Output:
	// {"propBool":false,"propInt":1234567878910111212,"propInt8":123,"propInt16":12349,"propInt32":1234567891,"propInt64":1234567878910111213,"propUint":12345678789101112138,"propUint8":255,"propUint16":12345,"propUint32":1234567891,"propUint64":12345678789101112139,"propFloat32":21.232426,"propFloat64":2799999999888.2827,"propString":"thirty two thirty four","propStruct":{"propName":["a name","another name","another"],"ps":["test pointer string",null,"test pointer string"]}}

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

//

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
