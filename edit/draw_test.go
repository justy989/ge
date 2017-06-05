package edit

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
	terminal_dimensions.X, terminal_dimensions.Y = termbox.Size()

	tabs := TabListLayout{}
	tabs.Tabs = append(tabs.Tabs, TabLayout{})
	current_tab := &tabs.Tabs[tabs.Selection]
	root_layout := ViewLayout{}
	root_layout.View.Buffer = buffer
	current_tab.Root = &root_layout
	current_tab.Selection = current_tab.Root

	settings := Settings{Draw: DrawSettings{4}}

	full_view := Rect{0, 0, terminal_dimensions.X, terminal_dimensions.Y}
	tabs.CalculateRect(full_view)

	for i := 0; i < b.N; i++ {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		tabs.Draw(terminal_dimensions, &settings.Draw)
		termbox.Flush()
	}
}
