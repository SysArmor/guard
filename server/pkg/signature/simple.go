package signature

import (
	"crypto/md5"
	"encoding/hex"
)

type simpleBody interface {
	String() string
}

type SimpleString string

func (s SimpleString) String() string {
	return string(s)
}

// SimpleSignature function with a generic type constraint
// TODO: maybe use bytes.Buffer instead of string concatenation
func SimpleSignature[T simpleBody](body T, key []byte) string {
	h := md5.New()
	h.Write([]byte(body.String()))
	return hex.EncodeToString(h.Sum(key))
}
