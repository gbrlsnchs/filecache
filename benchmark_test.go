package filecache_test

import (
	"testing"

	. "github.com/gbrlsnchs/filecache"
)

func BenchmarkReadDir(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := ReadDir("testdata", ""); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	c, err := ReadDir("testdata", "")
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Get("lorem.txt")
	}
}
