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
	"time"
)

// TODO: greetings 'something about a go pro'

func main() {
	flag.Parse()
	files := flag.Args()
	logfile, err := os.OpenFile("ge.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Fatalf("Open(ge.log) error: %v", err)
	}

	defer logfile.Close()
	log.SetOutput(logfile)

	var buffers []Buffer
	for _, file := range files {
		var f io.Reader
		f, err = os.Open(file)
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
		var b Buffer
		b = &BaseBuffer{}
		Load(b, f)
		b = NewUndoer(NewGoSyntaxHighlighter(b))
		buffers = append(buffers, b)
	}
	if len(buffers) == 0 {
		panic("Ahhh you didn't load any buffers. We should be able to handle this eventually")
	}

	err = termbox.Init()
	if err != nil {
		return
	}
	defer termbox.Close()

	terminal_dimensions := Point{}
	terminal_dimensions.x, terminal_dimensions.y = termbox.Size()

	tabs := TabListLayout{}
	tabs.tabs = append(tabs.tabs, TabLayout{})
	current_tab := &tabs.tabs[tabs.selection]
	root_layout := ViewLayout{}
	root_layout.view.buffer = buffers[0]
	current_tab.root = &root_layout
	current_tab.selection = current_tab.root

	// TODO: split layout with buffers that we loaded
	cursor_on_terminal := Point{0, 0}
	settings := Settings{draw: DrawSettings{4}}

	event_chan := make(chan termbox.Event, 1)
	go func() {
		for {
			event_chan <- termbox.PollEvent()
		}
	}()

	var vim Vim
	vim.init()

loop:
	for {
		terminal_dimensions.x, terminal_dimensions.y = termbox.Size()
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		full_view := Rect{0, 0, terminal_dimensions.x, terminal_dimensions.y}
		tabs.CalculateRect(full_view)
		tabs.Draw(terminal_dimensions, &settings.draw)
		selected_view_layout, selected_layout_is_view := current_tab.selection.(*ViewLayout)
		var b Buffer
		if selected_layout_is_view {
			b = selected_view_layout.view.buffer
			cursor_on_terminal = calc_cursor_on_terminal(
				PrintableCursor(b, b.Cursor(), &settings.draw),
				selected_view_layout.view.scroll,
				Point{selected_view_layout.view.rect.left, selected_view_layout.view.rect.top})
			termbox.SetCursor(cursor_on_terminal.x, cursor_on_terminal.y)
		}

		termbox.Flush()

		select {
		case ev := <-event_chan:
			switch ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEsc:
					break loop
				case termbox.KeyCtrlJ:
					current_tab.Select(DIRECTION_DOWN)
				case termbox.KeyCtrlK:
					current_tab.Select(DIRECTION_UP)
				case termbox.KeyCtrlH:
					current_tab.Select(DIRECTION_LEFT)
				case termbox.KeyCtrlL:
					current_tab.Select(DIRECTION_RIGHT)
				case termbox.KeyCtrlS:
					current_tab.Split()
				case termbox.KeyCtrlQ:
					current_tab.Remove()
				case termbox.KeyCtrlC:
					current_tab.Select(DIRECTION_IN)
				case termbox.KeyCtrlP:
					current_tab.Select(DIRECTION_OUT)
				case termbox.KeyCtrlB:
					current_tab.PrepareSplit(true)
				case termbox.KeyCtrlV:
					current_tab.PrepareSplit(false)
				case termbox.KeyCtrlN:
					list_layout, is_list_layout := current_tab.selection.(*ListLayout)
					if is_list_layout {
						list_layout.SetHorizontal(true)
						current_tab.CalculateRect(full_view)
					}
				case termbox.KeyCtrlM:
					list_layout, is_list_layout := current_tab.selection.(*ListLayout)
					if is_list_layout {
						list_layout.SetHorizontal(false)
						current_tab.CalculateRect(full_view)
					}
				case termbox.KeyCtrlT:
					new_tab := TabLayout{}
					new_view_layout := ViewLayout{}
					new_view_layout.view.buffer = buffers[0]
					new_tab.root = &new_view_layout
					new_tab.selection = new_tab.root
					tabs.tabs = append(tabs.tabs, new_tab)
				case termbox.KeyCtrlY:
					tabs.selection++
					tabs.selection %= len(tabs.tabs)
					current_tab = &tabs.tabs[tabs.selection]
				default:
					if selected_layout_is_view && b != nil {
						switch ev.Ch {
						default:
							state, action := vim.ParseAction(ev.Ch)
							if state == PARSE_ACTION_STATE_COMPLETE {
								err := vim.Perform(&action, b)
								if err == nil {
									//selected_view_layout.view.cursor = b.Cursor()
								} else {
									log.Println(err)
								}
							}
						case 'G':
							new_cursor := Point{0, len(b.Lines()) - 1}
							b.SetCursor(ClampOn(b, new_cursor))
						case '$':
							new_cursor := Point{len(b.Lines()[b.Cursor().y]) - 1, b.Cursor().y}
							b.SetCursor(ClampOn(b, new_cursor))
						case '0':
							new_cursor := Point{0, b.Cursor().y}
							b.SetCursor(ClampOn(b, new_cursor))
						case 'J':
							Join(b, b.Cursor().y)
						case 'u':
							undoer, ok := b.(Undoer)
							if ok {
								undoer.Undo()
							}
						case 'r':
							undoer, ok := b.(Undoer)
							if ok {
								undoer.Redo()
							}
						}
						selected_view_layout.view.cursor = b.Cursor()
					}
				}

				selected_view_layout, selected_layout_is_view = current_tab.selection.(*ViewLayout)
				if selected_layout_is_view {
					selected_view_layout.view.ScrollTo(
						PrintableCursor(selected_view_layout.view.buffer, selected_view_layout.view.buffer.Cursor(), &settings.draw))
				}

			}
		case <-time.After(time.Millisecond * 500):
			// re-draw every 500 milliseconds even if we didn't receive a keypress
		}
	}
}
