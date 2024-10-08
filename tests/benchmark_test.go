package main

import (
	"net/http"
	"testing"
)

const url = "http://localhost:9292/"

func BenchmarkLocalServer(b *testing.B) {
	b.ResetTimer() // exclude setup time
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(url)
		if err != nil {
			b.Fatalf("Failed to send request: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("Received non-200 response: %d", resp.StatusCode)
		}
	}
}

func BenchmarkLocalServerParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Get(url)
			if err != nil {
				b.Fatalf("Failed to send request: %v", err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				b.Fatalf("Received non-200 response: %d", resp.StatusCode)
			}
		}
	})
}

// run with: go test -bench=.
// notes:
// - this is synthetic benchmarking, not a real-world scenario. see wrk, ab, or hey for that
