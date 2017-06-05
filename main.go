package main

import (
	"compress/bzip2"
	"compress/gzip"
	"flag"
	"ge/edit"
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

	var buffers []edit.Buffer
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
		b := edit.NewUndoer(&edit.BaseBuffer{})
		edit.Load(b, f)
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

	terminal_dimensions := edit.Point{}
	terminal_dimensions.X, terminal_dimensions.Y = termbox.Size()

	tabs := edit.TabListLayout{}
	tabs.Tabs = append(tabs.Tabs, edit.TabLayout{})
	current_tab := &tabs.Tabs[tabs.Selection]
	root_layout := edit.ViewLayout{}
	root_layout.View.Buffer = buffers[0]
	current_tab.Root = &root_layout
	current_tab.Selection = current_tab.Root

	// TODO: split layout with buffers that we loaded
	cursor_on_terminal := edit.Point{0, 0}
	settings := edit.Settings{Draw: edit.DrawSettings{4}}

	event_chan := make(chan termbox.Event, 1)
	go func() {
		for {
			event_chan <- termbox.PollEvent()
		}
	}()

	var vim edit.Vim
	vim.Init()

loop:
	for {
		terminal_dimensions.X, terminal_dimensions.Y = termbox.Size()
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		full_view := edit.Rect{0, 0, terminal_dimensions.X, terminal_dimensions.Y}
		tabs.CalculateRect(full_view)
		tabs.Draw(terminal_dimensions, &settings.Draw)
		selected_view_layout, selected_layout_is_view := current_tab.Selection.(*edit.ViewLayout)
		var b edit.Buffer
		if selected_layout_is_view {
			b = selected_view_layout.View.Buffer
			cursor_on_terminal = edit.Calc_cursor_on_terminal(
				edit.PrintableCursor(b, b.Cursor(), &settings.Draw),
				selected_view_layout.View.Scroll,
				edit.Point{selected_view_layout.View.Rect.Left, selected_view_layout.View.Rect.Top})
			termbox.SetCursor(cursor_on_terminal.X, cursor_on_terminal.Y)
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
					current_tab.Select(edit.DIRECTION_DOWN)
				case termbox.KeyCtrlK:
					current_tab.Select(edit.DIRECTION_UP)
				case termbox.KeyCtrlH:
					current_tab.Select(edit.DIRECTION_LEFT)
				case termbox.KeyCtrlL:
					current_tab.Select(edit.DIRECTION_RIGHT)
				case termbox.KeyCtrlS:
					current_tab.Split()
				case termbox.KeyCtrlQ:
					current_tab.Remove()
				case termbox.KeyCtrlC:
					current_tab.Select(edit.DIRECTION_IN)
				case termbox.KeyCtrlP:
					current_tab.Select(edit.DIRECTION_OUT)
				case termbox.KeyCtrlB:
					current_tab.PrepareSplit(true)
				case termbox.KeyCtrlV:
					current_tab.PrepareSplit(false)
				case termbox.KeyCtrlN:
					list_layout, is_list_layout := current_tab.Selection.(*edit.ListLayout)
					if is_list_layout {
						list_layout.SetHorizontal(true)
						current_tab.CalculateRect(full_view)
					}
				case termbox.KeyCtrlM:
					list_layout, is_list_layout := current_tab.Selection.(*edit.ListLayout)
					if is_list_layout {
						list_layout.SetHorizontal(false)
						current_tab.CalculateRect(full_view)
					}
				case termbox.KeyCtrlT:
					new_tab := edit.TabLayout{}
					new_view_layout := edit.ViewLayout{}
					new_view_layout.View.Buffer = buffers[0]
					new_tab.Root = &new_view_layout
					new_tab.Selection = new_tab.Root
					tabs.Tabs = append(tabs.Tabs, new_tab)
				case termbox.KeyCtrlY:
					tabs.Selection++
					tabs.Selection %= len(tabs.Tabs)
					current_tab = &tabs.Tabs[tabs.Selection]
				default:
					if selected_layout_is_view && b != nil {
						switch ev.Ch {
						default:
							state, action := vim.ParseAction(ev.Ch)
							if state == edit.PARSE_ACTION_STATE_COMPLETE {
								err := vim.Perform(&action, b)
								if err == nil {
									//selected_view_layout.View.cursor = b.Cursor()
								} else {
									log.Println(err)
								}
							}
						case 'G':
							new_cursor := edit.Point{0, len(b.Lines()) - 1}
							b.SetCursor(edit.ClampOn(b, new_cursor))
						case '$':
							new_cursor := edit.Point{len(b.Lines()[b.Cursor().Y]) - 1, b.Cursor().Y}
							b.SetCursor(edit.ClampOn(b, new_cursor))
						case '0':
							new_cursor := edit.Point{0, b.Cursor().Y}
							b.SetCursor(edit.ClampOn(b, new_cursor))
						case 'J':
							edit.Join(b, b.Cursor().Y)
						case 'u':
							undoer, ok := b.(edit.Undoer)
							if ok {
								undoer.Undo()
							}
						case 'r':
							undoer, ok := b.(edit.Undoer)
							if ok {
								undoer.Redo()
							}
						}
						selected_view_layout.View.Cursor = b.Cursor()
					}
				}

				selected_view_layout, selected_layout_is_view = current_tab.Selection.(*edit.ViewLayout)
				if selected_layout_is_view {
					selected_view_layout.View.ScrollTo(
						edit.PrintableCursor(selected_view_layout.View.Buffer, selected_view_layout.View.Buffer.Cursor(), &settings.Draw))
				}

			}
		case <-time.After(time.Millisecond * 500):
			// re-draw every 500 milliseconds even if we didn't receive a keypress
		}
	}
}
