// artifact/bench_test.go
package artifact

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkGlob13k(b *testing.B) {
	dir := b.TempDir()
	for i := 0; i < 13000; i++ {
		subdir := filepath.Join(dir, fmt.Sprintf("d%d", i/100))
		os.MkdirAll(subdir, 0755)
		os.WriteFile(filepath.Join(subdir, fmt.Sprintf("f%d.cpp", i)), []byte("//"), 0644)
	}
	g := NewGlobSet(dir, "**/*.cpp")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Resolve(context.Background())
	}
}
