package main

import (
	"github.com/nsf/termbox-go"
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

// benchmark drawing buffer.go
func BenchmarkDrawBuffer(b *testing.B) {
	var f io.Reader
	f, err := os.Open("buffer.go")
	if err != nil {
		log.Fatalf("Open() error: %v", err)
	}
	defer f.(io.ReadCloser).Close()

	err = termbox.Init()
	if err != nil {
		return
	}
	defer termbox.Close()

	buffer := &BaseBuffer{}
	Load(buffer, f)

	terminal_dimensions := Point{}
	terminal_dimensions.x, terminal_dimensions.y = termbox.Size()

	tabs := TabListLayout{}
	tabs.tabs = append(tabs.tabs, TabLayout{})
	current_tab := &tabs.tabs[tabs.selection]
	root_layout := ViewLayout{}
	root_layout.view.buffer = buffer
	current_tab.root = &root_layout
	current_tab.selection = current_tab.root

	settings := Settings{draw: DrawSettings{4}}

	full_view := Rect{0, 0, terminal_dimensions.x, terminal_dimensions.y}
	tabs.CalculateRect(full_view)

	for i := 0; i < b.N; i++ {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		tabs.Draw(terminal_dimensions, &settings.draw)
		termbox.Flush()
	}
}
