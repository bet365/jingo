package jingo

import (
	"bytes"
	"encoding"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestMapEncoderUnsupportedTypeError(t *testing.T) {

	tests := []struct {
		name string
		t    interface{}
		want string
	}{
		{
			"unsupported key type: struct",
			map[struct{}]string{},
			"unsupported key type",
		},
		{
			"unsupported elem type: chan",
			map[string]chan string{},
			"unsupported elem type",
		},
		{
			"unsupported elem type: chan",
			map[string]*chan string{},
			"unsupported ptr elem type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				v := recover().(string)

				if v != tt.want {
					t.Fatalf("\nWanted:\n%q\nGot:\n%q", tt.want, v)
				}
			}()
			NewMapEncoder(tt.t)
		})
	}
}

type textStruct struct {
	text []byte
}

func (t textStruct) MarshalText() ([]byte, error) { return t.text, nil }

var _ encoding.TextMarshaler = &textStruct{}

func TestMapEncoder_key_marshaltext(t *testing.T) {

	enc := NewMapEncoder(map[*textStruct]string{})

	tests := []struct {
		name string
		v    map[*textStruct]string
		want []byte
	}{
		{
			"Nil",
			map[*textStruct]string(nil),
			[]byte(`null`),
		},
		{
			"Empty",
			map[*textStruct]string{},
			[]byte(`{}`),
		},
		{
			"Nil Key",
			map[*textStruct]string{{nil}: "aa"},
			[]byte(`{"":"aa"}`),
		},
		{
			"Non-nil key",
			map[*textStruct]string{{[]byte("1")}: "aa"},
			[]byte(`{"1":"aa"}`),
		},
		{
			"Many",
			map[*textStruct]string{
				{[]byte("2")}: "aa",
				{[]byte("3")}: "bb",
				{[]byte("1")}: "cc",
				{nil}:         "dd",
			},
			[]byte(`{"":"dd","1":"cc","2":"aa","3":"bb"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_key_time(t *testing.T) {

	enc := NewMapEncoder(map[time.Time]string{})

	d0 := time.Date(2000, 9, 17, 20, 4, 26, 0, time.UTC)
	d1 := time.Date(2001, 9, 17, 20, 4, 26, 0, time.UTC)
	d2 := time.Date(2002, 9, 17, 20, 4, 26, 0, time.UTC)

	tests := []struct {
		name string
		v    map[time.Time]string
		want []byte
	}{
		{
			"One",
			map[time.Time]string{
				d0: "1",
			},
			[]byte(`{"2000-09-17T20:04:26Z":"1"}`),
		},
		{
			"Many",
			map[time.Time]string{
				d0: "1",
				d1: "2",
				d2: "3",
			},
			[]byte(`{"2000-09-17T20:04:26Z":"1","2001-09-17T20:04:26Z":"2","2002-09-17T20:04:26Z":"3"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_key_ptrtime(t *testing.T) {

	enc := NewMapEncoder(map[*time.Time]string{})

	d0 := time.Date(2000, 9, 17, 20, 4, 26, 0, time.UTC)
	d1 := time.Date(2001, 9, 17, 20, 4, 26, 0, time.UTC)
	d2 := time.Date(2002, 9, 17, 20, 4, 26, 0, time.UTC)

	tests := []struct {
		name string
		v    map[*time.Time]string
		want []byte
	}{
		{
			"Nil",
			map[*time.Time]string{
				nil: "1",
			},
			[]byte(`{"":"1"}`),
		},
		{
			"Non-nil",
			map[*time.Time]string{
				&d0: "1",
			},
			[]byte(`{"2000-09-17T20:04:26Z":"1"}`),
		},
		{
			"Many",
			map[*time.Time]string{
				nil: "1",
				&d0: "2",
				&d1: "3",
				&d2: "4",
			},
			[]byte(`{"":"1","2000-09-17T20:04:26Z":"2","2001-09-17T20:04:26Z":"3","2002-09-17T20:04:26Z":"4"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

type innerStruct struct {
	S string `json:"s"`
}

func TestMapEncoder_elem_struct(t *testing.T) {

	enc := NewMapEncoder(map[string]innerStruct{})

	tests := []struct {
		name string
		v    map[string]innerStruct
		want []byte
	}{
		{
			"One",
			map[string]innerStruct{"a": {"s0"}},
			[]byte(`{"a":{"s":"s0"}}`),
		},
		{
			"Many",
			map[string]innerStruct{"a": {"s1"}, "b": {"s2"}, "c": {"s3"}},
			[]byte(`{"a":{"s":"s1"},"b":{"s":"s2"},"c":{"s":"s3"}}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_elem_ptrstruct(t *testing.T) {

	enc := NewMapEncoder(map[string]*innerStruct{})

	var (
		s0 = innerStruct{"s0"}
		s1 = innerStruct{"s1"}
	)

	tests := []struct {
		name string
		v    map[string]*innerStruct
		want []byte
	}{
		{
			"Nil",
			map[string]*innerStruct{"a": nil},
			[]byte(`{"a":null}`),
		},
		{
			"One",
			map[string]*innerStruct{
				"a": &s0,
			},
			[]byte(`{"a":{"s":"s0"}}`),
		},
		{
			"Many",
			map[string]*innerStruct{
				"a": &s0,
				"b": &s1,
				"c": nil,
			},
			[]byte(`{"a":{"s":"s0"},"b":{"s":"s1"},"c":null}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_elem_slice(t *testing.T) {

	enc := NewMapEncoder(map[string][]int{})

	tests := []struct {
		name string
		v    map[string][]int
		want []byte
	}{
		{
			"nil",
			map[string][]int{"a": nil},
			[]byte(`{"a":[]}`),
		},
		{
			"One",
			map[string][]int{"a": {1, 2, 3}},
			[]byte(`{"a":[1,2,3]}`),
		},
		{
			"Many",
			map[string][]int{"a": {1, 2, 3}, "b": {4, 5, 6}, "c": nil},
			[]byte(`{"a":[1,2,3],"b":[4,5,6],"c":[]}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_elem_ptrslice(t *testing.T) {

	enc := NewMapEncoder(map[string]*[]int{})

	var (
		s0 = []int{1, 2, 3}
		s1 = []int{4, 5, 6}
	)

	tests := []struct {
		name string
		v    map[string]*[]int
		want []byte
	}{
		{
			"nil",
			map[string]*[]int{"a": nil},
			[]byte(`{"a":null}`),
		},
		{
			"One",
			map[string]*[]int{"a": &s0},
			[]byte(`{"a":[1,2,3]}`),
		},
		{
			"Many",
			map[string]*[]int{"a": &s0, "b": &s1, "c": nil},
			[]byte(`{"a":[1,2,3],"b":[4,5,6],"c":null}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_elem_map(t *testing.T) {

	enc := NewMapEncoder(map[string]map[string]string{})

	tests := []struct {
		name string
		v    map[string]map[string]string
		want []byte
	}{
		{
			"Nil",
			map[string]map[string]string(nil),
			[]byte(`null`),
		},
		{
			"Empty",
			map[string]map[string]string{},
			[]byte(`{}`),
		},
		{
			"One - Nil Elem",
			map[string]map[string]string{"1": nil},
			[]byte(`{"1":null}`),
		},
		{
			"One",
			map[string]map[string]string{"1": {"KA": "EA"}},
			[]byte(`{"1":{"KA":"EA"}}`),
		},
		{
			"Many",
			map[string]map[string]string{
				"1": {"KA": "EA"},
				"2": {"KB1": "EB1", "KB2": "EB2", "KB3": "EK3"},
				"3": {"KC1": "EC1", "KC2": "EC2"},
			},
			[]byte(`{"1":{"KA":"EA"},"2":{"KB1":"EB1","KB2":"EB2","KB3":"EK3"},"3":{"KC1":"EC1","KC2":"EC2"}}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_elem_ptrmap(t *testing.T) {

	enc := NewMapEncoder(map[string]*map[string]string{})

	var (
		m1 = map[string]string{"KA1": "EA1"}
		m2 = map[string]string{"KB1": "EB1", "KB2": "EB2", "KB3": "EK3"}
		m3 = map[string]string{"KC1": "EC1", "KC2": "EC2"}
	)

	tests := []struct {
		name string
		v    map[string]*map[string]string
		want []byte
	}{
		{
			"One - Nil Elem",
			map[string]*map[string]string{"1": nil},
			[]byte(`{"1":null}`),
		},
		{
			"One",
			map[string]*map[string]string{"1": &m1},
			[]byte(`{"1":{"KA1":"EA1"}}`),
		},
		{
			"Many",
			map[string]*map[string]string{
				"1": &m1,
				"2": &m2,
				"3": &m3,
				"4": nil,
			},
			[]byte(`{"1":{"KA1":"EA1"},"2":{"KB1":"EB1","KB2":"EB2","KB3":"EK3"},"3":{"KC1":"EC1","KC2":"EC2"},"4":null}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_elem_ptrstring(t *testing.T) {

	enc := NewMapEncoder(map[string]*string{})

	var (
		aa = "aa"
		bb = "bb"
		cc = "cc"
	)

	tests := []struct {
		name string
		v    map[string]*string
		want []byte
	}{
		{
			"Nil",
			map[string]*string(nil),
			[]byte(`null`),
		},
		{
			"Empty",
			map[string]*string{},
			[]byte(`{}`),
		},
		{
			"One - Nil",
			map[string]*string{"1": nil},
			[]byte(`{"1":null}`),
		},
		{
			"One",
			map[string]*string{"2": &aa},
			[]byte(`{"2":"aa"}`),
		},
		{
			"Many - Mixed",
			map[string]*string{"3": nil, "2": &cc, "1": nil},
			[]byte(`{"1":null,"2":"cc","3":null}`),
		},
		{
			"Many",
			map[string]*string{"3": &aa, "1": &bb, "2": &cc},
			[]byte(`{"1":"bb","2":"cc","3":"aa"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_elem_nonstring(t *testing.T) {

	enc := NewMapEncoder(map[string]int{})

	tests := []struct {
		name string
		v    map[string]int
		want []byte
	}{
		{
			"One",
			map[string]int{"2": 1},
			[]byte(`{"2":1}`),
		},
		{
			"Many",
			map[string]int{"3": 1, "1": 2, "2": 3},
			[]byte(`{"1":2,"2":3,"3":1}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}
func TestMapEncoder_elem_ptrnonstring(t *testing.T) {

	enc := NewMapEncoder(map[string]*int{})

	var (
		one   = 1
		two   = 2
		three = 3
	)

	tests := []struct {
		name string
		v    map[string]*int
		want []byte
	}{
		{
			"One - Nil",
			map[string]*int{"2": nil},
			[]byte(`{"2":null}`),
		},
		{
			"One",
			map[string]*int{"2": &one},
			[]byte(`{"2":1}`),
		},
		{
			"Many - Mixed",
			map[string]*int{"3": nil, "2": &three, "1": nil},
			[]byte(`{"1":null,"2":3,"3":null}`),
		},
		{
			"Many",
			map[string]*int{"3": &one, "1": &two, "2": &three},
			[]byte(`{"1":2,"2":3,"3":1}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}
func TestMapEncoder_elem_marshaltext(t *testing.T) {

	enc := NewMapEncoder(map[string]textStruct{})

	tests := []struct {
		name string
		v    map[string]textStruct
		want []byte
	}{
		{
			"One - Nil",
			map[string]textStruct{"1": {nil}},
			[]byte(`{"1":""}`),
		},
		{
			"One",
			map[string]textStruct{"1": {[]byte("aa")}},
			[]byte(`{"1":"aa"}`),
		},
		{
			"Many",
			map[string]textStruct{"1": {[]byte("aa")}, "2": {[]byte("bb")}, "3": {[]byte("cc")}, "4": {nil}},
			[]byte(`{"1":"aa","2":"bb","3":"cc","4":""}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_elem_ptrmarshaltext(t *testing.T) {

	enc := NewMapEncoder(map[string]*textStruct{})

	aa := &textStruct{[]byte("aa")}

	tests := []struct {
		name string
		v    map[string]*textStruct
		want []byte
	}{
		{
			"One - Nil Elem",
			map[string]*textStruct{"1": nil},
			[]byte(`{"1":null}`),
		},
		{
			"One - Nil Elem Value",
			map[string]*textStruct{"1": nil},
			[]byte(`{"1":null}`),
		},
		{
			"One",
			map[string]*textStruct{"1": aa},
			[]byte(`{"1":"aa"}`),
		},
		{
			"Many",
			map[string]*textStruct{
				"1": {[]byte("aa")},
				"2": {[]byte("bb")},
				"3": {[]byte("cc")},
				"4": {nil},
				"5": nil,
			},
			[]byte(`{"1":"aa","2":"bb","3":"cc","4":"","5":null}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			enc.Marshal(&tt.v, buf)

			if !bytes.Equal(tt.want, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s\n", tt.want, buf.Bytes)
			}
		})
	}
}

type marshaler interface {
	Marshal(s interface{}, w *Buffer)
}

func TestMapEncoder_sorted_nonstring(t *testing.T) {

	tests := []struct {
		name string
		enc  marshaler
		v    interface{}
	}{
		{
			"key: int",
			NewMapEncoder(map[int]string{}),
			&map[int]string{
				4:        "A",
				59:       "B",
				238:      "C",
				-784:     "D",
				9845:     "E",
				959:      "F",
				905:      "G",
				0:        "H",
				42:       "I",
				7586:     "J",
				-5467984: "K",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()

			tt.enc.Marshal(tt.v, buf)

			bs, err := json.Marshal(tt.v)
			if err != nil {
				t.Fatalf("Unable to marshal value %s", err)
			}

			if !bytes.Equal(bs, buf.Bytes) {
				t.Errorf("\nwant:\n%s\ngot:\n%s", bs, buf.Bytes)
			}
		})
	}
}

func TestMapEncoder_unsorted_fast_string(t *testing.T) {

	var cfg Config
	cfg.SetSortMapKeys(false)

	enc := NewMapEncoderWithConfig(map[string]string{}, cfg)

	tests := []struct {
		name string
		v    map[string]string
	}{
		{
			"Nil",
			map[string]string(nil),
		},
		{
			"Empty",
			map[string]string{},
		},
		{
			"One",
			map[string]string{"aa": "123"},
		},
		{
			"Many",
			map[string]string{"bb": "678", "cc": "345", "aa": "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()
			enc.Marshal(&tt.v, buf)

			var dst map[string]string

			if err := json.Unmarshal(buf.Bytes, &dst); err != nil {
				t.Fatalf("Unable to unmarshal buf.Bytes -  %s", err)
			}

			if !reflect.DeepEqual(tt.v, dst) {
				t.Fatalf("buf.Bytes=%q\nWant:%+v\nGot:%+v", buf.Bytes, tt.v, dst)
			}
		})
	}
}

func TestMapEncoder_unsorted_non_string(t *testing.T) {

	var cfg Config
	cfg.SetSortMapKeys(false)

	enc := NewMapEncoderWithConfig(map[int]string{}, cfg)

	tests := []struct {
		name string
		v    map[int]string
	}{
		{
			"Nil",
			map[int]string(nil),
		},
		{
			"Empty",
			map[int]string{},
		},
		{
			"One",
			map[int]string{2: "aa"},
		},
		{
			"Many",
			map[int]string{3: "cc", 1: "bb", 2: "aa", 4: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := NewBufferFromPool()
			defer buf.ReturnToPool()
			enc.Marshal(&tt.v, buf)

			var dst map[int]string

			if err := json.Unmarshal(buf.Bytes, &dst); err != nil {
				t.Fatalf("Unable to unmarshal buf.Bytes -  %s\nbuf.Bytes=%s", err, buf.Bytes)
			}

			if !reflect.DeepEqual(tt.v, dst) {
				t.Fatalf("\nWant:%+v\nGot:%+v\nbuf.Bytes=%s", tt.v, dst, buf.Bytes)
			}
		})
	}
}
