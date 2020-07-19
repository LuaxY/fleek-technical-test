package store

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewMemoryStore(t *testing.T) {
	ms, err := NewMemoryStore()

	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		key     string
		value   interface{}
		ret     interface{}
		compare func(a interface{}, b interface{}) bool
	}{
		{
			key:   "int",
			value: 42,
			ret:   0,
			compare: func(a interface{}, b interface{}) bool {
				aInt := a.(int)
				bInt := b.(int)
				return aInt == bInt
			},
		},
		{
			key:   "bytes",
			value: []byte("bytes"),
			ret:   []byte{},
			compare: func(a interface{}, b interface{}) bool {
				aBytes := a.([]byte)
				bBytes := b.([]byte)
				return bytes.Compare(aBytes, bBytes) == 0
			},
		},
		{
			key:   "string",
			value: "string",
			ret:   "",
			compare: func(a interface{}, b interface{}) bool {
				aString := a.(string)
				bString := b.(string)
				return strings.Compare(aString, bString) == 0
			},
		},
	}

	for _, test := range testCases {
		ms.Add(test.key, test.value)
		ms.Get(test.key, &test.ret)

		if !test.compare(test.ret, test.value) {
			t.Fatalf("'%v' must be equal to '%v'", test.ret, test.value)
		}
	}
}
