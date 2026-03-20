// fingerprint/bench_test.go
package fingerprint

import (
	"fmt"
	"testing"
)

func BenchmarkMerkleRoot13k(b *testing.B) {
	fps := make([]string, 13000)
	for i := range fps {
		fps[i] = fmt.Sprintf("fp_%d_aaabbbccc", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MerkleRoot(fps)
	}
}
