package idx

import (
	"testing"
	"time"
)

func BenchmarkSnowflake_NextBatch(b *testing.B) {
	idg := NewSnowflakeWithPool(1, 5000000, 100000, func() int64 {
		return time.Now().UnixMilli()
	})
	for i := 0; i < b.N; i++ {
		idg.Next()
	}
}

func TestSnowflake_NextBatch(t *testing.T) {
	idg := NewSnowflakeWithPool(1, 1000000, 5000, func() int64 { return 1764840058824 })

	cases := []struct {
		// got  int
		want int
	}{
		{651714557688942592},
		{651714557688942593},
		{651714557688942594},
		{651714557688942595},
	}

	for _, c := range cases {
		if idg.Next() != int64(c.want) {
			t.Errorf("Next() = %v want %v", idg.Next(), c.want)
		}
	}
}
