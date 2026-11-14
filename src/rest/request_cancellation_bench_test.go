package rest

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func BenchmarkMultipartEncoding(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, size := range []int{1024, 1 << 20} {
		b.Run(fmt.Sprintf("%d-bytes", size), func(b *testing.B) {
			data := strings.Repeat("x", size)
			reader := strings.NewReader(data)
			r := &request{body: map[string]string{"content": "fixture"}, files: []File{{Name: "fixture.txt", Reader: reader}}}
			b.ReportAllocs()
			b.SetBytes(int64(size))
			b.ResetTimer()
			for range b.N {
				reader.Reset(data)
				if _, err := r.encode(ctx); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
