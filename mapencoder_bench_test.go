package jingo

import (
	"testing"
)

func BenchmarkMapEncoder(b *testing.B) {

	cfg := DefaultConfig()
	cfg.SetSortMapKeys(true)

	for _, bb := range []struct {
		name    string
		encoder marshaler
		s       interface{}
	}{
		{
			"Key: string, Elem: string",
			NewMapEncoder(map[string]string{}),
			&map[string]string{
				"10": "1",
				"9":  "2",
				"8":  "3",
				"7":  "4",
				"6":  "5",
				"5":  "6",
				"4":  "7",
				"3":  "8",
				"2":  "9",
				"1":  "10",
			},
		},
		{
			"Key: string, Elem: string sorted",
			NewMapEncoderWithConfig(map[string]string{}, cfg),
			&map[string]string{
				"10": "1",
				"9":  "2",
				"8":  "3",
				"7":  "4",
				"6":  "5",
				"5":  "6",
				"4":  "7",
				"3":  "8",
				"2":  "9",
				"1":  "10",
			},
		},
		{
			"Key: int, Elem: string",
			NewMapEncoder(map[int]string{}),
			&map[int]string{
				10: "1",
				9:  "2",
				8:  "3",
				7:  "4",
				6:  "5",
				5:  "6",
				4:  "7",
				3:  "8",
				2:  "9",
				1:  "10",
			},
		},
		{
			"Key: int, Elem: string sorted",
			NewMapEncoderWithConfig(map[int]string{}, cfg),
			&map[int]string{
				10: "1",
				9:  "2",
				8:  "3",
				7:  "4",
				6:  "5",
				5:  "6",
				4:  "7",
				3:  "8",
				2:  "9",
				1:  "10",
			},
		},
		{
			"Key: MarshalText, Elem: string",
			NewMapEncoder(map[*textStruct]string{}),
			&map[*textStruct]string{
				{[]byte("10")}: "1",
				{[]byte("9")}:  "2",
				{[]byte("8")}:  "3",
				{[]byte("7")}:  "4",
				{[]byte("6")}:  "5",
				{[]byte("5")}:  "6",
				{[]byte("4")}:  "7",
				{[]byte("3")}:  "8",
				{[]byte("2")}:  "9",
				{[]byte("1")}:  "10",
			},
		},
		{
			"Key: MarshalText, Elem: string sorted",
			NewMapEncoderWithConfig(map[*textStruct]string{}, cfg),
			&map[*textStruct]string{
				{[]byte("10")}: "1",
				{[]byte("9")}:  "2",
				{[]byte("8")}:  "3",
				{[]byte("7")}:  "4",
				{[]byte("6")}:  "5",
				{[]byte("5")}:  "6",
				{[]byte("4")}:  "7",
				{[]byte("3")}:  "8",
				{[]byte("2")}:  "9",
				{[]byte("1")}:  "10",
			},
		},
	} {
		b.Run(bb.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {

				buf := NewBufferFromPool()

				bb.encoder.Marshal(bb.s, buf)

				buf.ReturnToPool()
			}
		})
	}
}
