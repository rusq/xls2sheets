package authmgr

import (
	"fmt"
	"testing"
)

func Benchmark_randString(b *testing.B) {
	stringSz := []int{1, 2, 4, 8, 16, 32, 64}
	for _, sz := range stringSz {
		b.Run(fmt.Sprintf("sz:%d", sz), func(b *testing.B) {
			for i := 0; i <= b.N; i++ {
				randString(sz)
			}
		})
	}
}
