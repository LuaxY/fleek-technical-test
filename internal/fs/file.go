package fs

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/pkg/errors"
)

type File struct {
	path string
	hash string
	key  []byte
	name string
	size int64
}

func NewFile(path string, reader io.Reader, name string, size int64) (*File, error) {
	f := File{
		path: path,
	}

	// to have a unique hash for each files, regardless of content, because multiple files can have
	// same content, we concatenate filename and content to be sure we don't have conflicts.
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(name))
	if _, err := io.Copy(hasher, reader); err != nil {
		return nil, errors.Wrap(err, "compute sha-256 of file")
	}

	f.hash = hex.EncodeToString(hasher.Sum(nil))
	f.name = name
	f.size = size

	key := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, errors.Wrap(err, "generate random key")
	}

	f.key = key

	return &f, nil
}

func (f *File) Path() string {
	return f.path
}

func (f *File) Hash() string {
	return f.hash
}

func (f *File) Key() []byte {
	return f.key
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Size() int64 {
	return f.size
}

func EncryptDecrypt(key []byte, src io.Reader, dst io.Writer) error {
	block, err := aes.NewCipher(key)

	if err != nil {
		return errors.Wrap(err, "initiate new aes cipher")
	}

	// since key is unique for each file, it's ok to use a zero IV
	var iv [aes.BlockSize]byte
	stream := cipher.NewCTR(block, iv[:])
	writer := &cipher.StreamWriter{S: stream, W: dst}

	if _, err := io.Copy(writer, src); err != nil {
		return errors.Wrap(err, "aes ctr encryption")
	}

	return nil
}
