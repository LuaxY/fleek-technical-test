package fs

import (
	"bytes"
	"testing"
)

func TestFile_EncryptDecrypt(t *testing.T) {
	testCases := [][]byte{
		[]byte(""), // empty
		[]byte("text"),
		[]byte("some text with space"),
		[]byte("']'&$']' '//!']'# $[]D&(!4# (#41²"),
		[]byte("ⓣⓔⓢⓣ ⓦⓘⓣⓗ ⓢⓟⓔⓒⓘⓐⓛ ⓒⓗⓐⓡ"),
	}

	for _, data := range testCases {
		in := bytes.NewReader(data)

		file, err := NewFile("path/to/file", &bytes.Buffer{}, "file", 1234)

		if err != nil {
			t.Fatal(err)
		}

		var encrypted bytes.Buffer

		if err = EncryptDecrypt(file.Key(), in, &encrypted); err != nil {
			t.Fatal(err)
		}

		var decrypted bytes.Buffer

		if err = EncryptDecrypt(file.Key(), &encrypted, &decrypted); err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(decrypted.Bytes(), data) != 0 {
			t.Fatalf("'%s' is not equal to '%s'\nkey: %x", decrypted.String(), data, file.key)
		}
	}
}
