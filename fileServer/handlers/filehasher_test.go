package handlers

import (
	"testing"
)

const (
	testPath = "/Users/ginahuang/tmp"
)

func BenchmarkConstructFileHashesBounded(b *testing.B) {
	for i := 0; i < b.N; i++ {
		constructFileHashesBounded(testPath)
	}
}

func BenchmarkConstructFileHashesUnbounded(b *testing.B) {
	for i := 0; i < b.N; i++ {
		constructFileHashesUnbounded(testPath)
	}
}
