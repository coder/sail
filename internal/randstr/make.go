// +build linux darwin freebsd

package randstr

import (
	"crypto/rand"
	"encoding/binary"
	"strings"
)

// The provided charsets.
const (
	Default = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	Alpha   = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	Upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Lower   = "abcdefghijklmnopqrstuvwxyz"
	Numeric = "0123456789"
	Hex     = "0123456789abcdef"
	// Human creates strings which are easily distinguishable from others
	// created with the same charset. It contains most lowercase alphanumeric characters without
	// 0,o,i,1,l.
	Human = "23456789abcdefghjkmnpqrstuvwxyz"
)

// Make returns a random string using Default.
func Make(size int) string {
	return MakeCharset(Default, size)
}

// MakeCharset generates a random string using the provided charset and size.
func MakeCharset(charsetStr string, size int) string {
	charset := []rune(charsetStr)

	// This buffer facilitates pre-emptively reading random uint32s
	// to reduce syscall overhead.
	ibuf := bytes(4 * size)

	var s strings.Builder
	s.Grow(size)
	for i := 0; i < size; i++ {
		// There is a bias in this we can improve on later.
		// See https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
		n := binary.BigEndian.Uint32(ibuf[i*4 : (i+1)*4])
		s.WriteRune(charset[int(n)%(len(charset))])
	}
	return s.String()
}

func must(err error) {
	if err != nil {
		panic("randstr: " + err.Error())
	}
}

func bytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	must(err)
	return b
}
