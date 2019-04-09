package randstr

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			str := Make(5)
			assert.Equal(t, 5, utf8.RuneCountInString(str))
		}
	})
	t.Run("Runes", func(t *testing.T) {
		charset := "💓😘💓🌷"
		for i := 0; i < 10; i++ {
			str := MakeCharset(charset, 10)
			assert.Equal(t, 10, utf8.RuneCountInString(str))
		}

	})
}

func BenchmarkMake20(b *testing.B) {
	b.SetBytes(20)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Make(20)
	}
}
