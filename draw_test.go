package main

import (
	"io"
	"log"
	"os"
	"testing"
)

// benchmark syntax highlighting buffer.go
func BenchmarkNewHighlighter(b *testing.B) {
	var f io.Reader
	f, err := os.Open("buffer.go")
	if err != nil {
		log.Fatalf("Open() error: %v", err)
	}
	defer f.(io.ReadCloser).Close()
	buffer := &BaseBuffer{}
	Load(buffer, f)
	for i := 0; i < b.N; i++ {
		highlighter := NewHighlighter(buffer)
		highlighter.Highlight(Point{0, 0})
	}
}
