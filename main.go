package main

import (
	"compress/bzip2"
	"compress/gzip"
	"flag"
	"github.com/nsf/termbox-go"
	"io"
	"log"
	"os"
	"strings"
)

func calc_cursor_on_terminal(cursor Point, scroll Point, view_top_left Point) Point {
	cursor.x = cursor.x - scroll.x + view_top_left.x
	cursor.y = cursor.y - scroll.y + view_top_left.y
	return cursor
}

func main() {
	flag.Parse()
	files := flag.Args()
	var buffers []*EditableBuffer
	for _, file := range files {
		var f io.Reader
		f, err := os.Open(file)
		if err != nil {
			log.Fatalf("Open() error: %v", err)
		}
		defer f.(io.ReadCloser).Close()

		// handle reading compressed files
		switch {
		case strings.HasSuffix(file, ".gz"):
			{
				f, err = gzip.NewReader(f)
				if err != nil {
					log.Fatalf("gzip error: %v", err)
				}
				defer f.(io.ReadCloser).Close()
			}
		case strings.HasSuffix(file, ".bz2"):
			{
				f = bzip2.NewReader(f)
				if err != nil {
					log.Fatalf("bzip error: %v", err)
				}
			}
		}

		log.Print("Loading " + file)
		b := NewEditableBuffer(&BaseBuffer{})
		b.Load(f)
		buffers = append(buffers, b)
	}
	if len(buffers) == 0 {
		panic("Ahhh you didn't load any buffers. We should be able to handle this eventually")
	}

	err := termbox.Init()
	if err != nil {
		return
	}
	defer termbox.Close()

	terminal_dimensions := Point{}
	terminal_dimensions.x, terminal_dimensions.y = termbox.Size()

	layout := VerticalLayout{}
	b := buffers[0]
	for _, buf := range buffers {
		layout.layouts = append(layout.layouts, &ViewLayout{View{buffer: buf}})
	}
	selected_layout := layout.layouts[0]

loop:
	for {
		terminal_dimensions.x, terminal_dimensions.y = termbox.Size()
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		full_view := Rect{0, 0, terminal_dimensions.x, terminal_dimensions.y}
		layout.CalculateView(full_view)
		layout.Draw(terminal_dimensions)
		view_selected_layout := selected_layout.(*ViewLayout)
		cursor_on_terminal := calc_cursor_on_terminal(view_selected_layout.view.cursor, view_selected_layout.view.scroll,
			Point{view_selected_layout.view.rect.left, view_selected_layout.view.rect.top})
		termbox.SetCursor(cursor_on_terminal.x, cursor_on_terminal.y)
		termbox.Flush()

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break loop
			case termbox.KeyCtrlJ:
				new_selected_layout := layout.Find(Point{view_selected_layout.view.buffer.Cursor().x, view_selected_layout.view.rect.bottom + 2})
				if new_selected_layout != nil {
					selected_layout = new_selected_layout
				}
			case termbox.KeyCtrlK:
				new_selected_layout := layout.Find(Point{view_selected_layout.view.buffer.Cursor().x, view_selected_layout.view.rect.top - 2})
				if new_selected_layout != nil {
					selected_layout = new_selected_layout
				}
			case termbox.KeyCtrlV:
				layout.SplitLayout(view_selected_layout)
			case termbox.KeyCtrlQ:
				if len(layout.layouts) > 1 {
					layout.Remove(selected_layout)
					layout.CalculateView(full_view)
					selected_layout = layout.Find(cursor_on_terminal)
				}
			default:
				switch ev.Ch {
				case 'h':
					view_selected_layout.view.cursor = b.MoveCursor(view_selected_layout.view.cursor, Point{-1, 0})
				case 'l':
					view_selected_layout.view.cursor = b.MoveCursor(view_selected_layout.view.cursor, Point{1, 0})
				case 'k':
					view_selected_layout.view.cursor = b.MoveCursor(view_selected_layout.view.cursor, Point{0, -1})
				case 'j':
					view_selected_layout.view.cursor = b.MoveCursor(view_selected_layout.view.cursor, Point{0, 1})
				case 'G':
					view_selected_layout.view.cursor = Point{0, len(b.Lines()) - 1}
					view_selected_layout.view.cursor = b.ClampOn(view_selected_layout.view.cursor)
				case '$':
					view_selected_layout.view.cursor = Point{len(b.Lines()[b.Cursor().y]) - 1, b.Cursor().y}
					view_selected_layout.view.cursor = b.ClampOn(view_selected_layout.view.cursor)
				case '0':
					view_selected_layout.view.cursor = Point{0, b.Cursor().y}
					view_selected_layout.view.cursor = b.ClampOn(view_selected_layout.view.cursor)
				}
			}

			view_selected_layout.view.ScrollTo(view_selected_layout.view.cursor)
		}
	}
}
