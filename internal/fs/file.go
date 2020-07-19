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

// File representation of encrypted file stored on disk with associated key and metadata like name, path, size, hash.
type File struct {
	path string
	hash string
	key  []byte
	name string
	size int64
}

// NewFile creates new representation of encrypted file by generating unique id using SHA-256 of concatenation of
// file content and relative file path. An AES 256 bits encryption key is also randomly generated.
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

// Path returns path of original file in source directory
func (f *File) Path() string {
	return f.path
}

// Hash returns SHA-256 hash of content file + relative path
func (f *File) Hash() string {
	return f.hash
}

// Key returns AES 256 bits encryption key
func (f *File) Key() []byte {
	return f.key
}

// Name returns relative file path
func (f *File) Name() string {
	return f.name
}

// Size returns size of file in bytes
func (f *File) Size() int64 {
	return f.size
}

// EncryptDecrypt encrypts or decrypt src to dst using provided key with AES CTR algorithm
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
