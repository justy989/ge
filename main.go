package main

import (
     "os"
     "github.com/nsf/termbox-go"
)

func (buffer* Buffer) Draw(view Rect, scroll Point, terminal_dimensions Point) {
     last_row := scroll.y + view.Height()
     if last_row > len(buffer.lines) {
          last_row = len(buffer.lines)
     }

     max_col := scroll.x + view.Width()
     for y, line := range buffer.lines[scroll.y:last_row] {
          if y >= view.Height() {
               break
          }

          last_col := max_col
          if last_col > len(line) {
               last_col = len(line)
          }

          if scroll.x > last_col {
               continue
          }

          final_y := y + view.top
          if final_y >= terminal_dimensions.y {
               break
          }

          for x, ch := range line[scroll.x:last_col] {
               if x >= view.Width() {
                    break
               }

               final_x := x + view.left
               if final_x >= terminal_dimensions.x {
                    break
               }

               termbox.SetCell(final_x, final_y, ch, termbox.ColorDefault, termbox.ColorDefault)
          }
     }
}

func calc_cursor_on_terminal(cursor Point, scroll Point, view_top_left Point) (Point) {
     cursor.x = cursor.x - scroll.x + view_top_left.x
     cursor.y = cursor.y - scroll.y + view_top_left.y
     return cursor
}

func main() {
     f, err := os.Open("main.go")
     if err != nil {
          return
     }

     b := NewFileBuffer(f)

     err = termbox.Init()
     if err != nil {
          return
     }
     defer termbox.Close()

     terminal_dimensions := Point{}
     terminal_dimensions.x, terminal_dimensions.y = termbox.Size()

     layout := VerticalLayout {}
     layout.layouts = append(layout.layouts, &ViewLayout{View{buffer: b}})
     layout.layouts = append(layout.layouts, &ViewLayout{View{buffer: b}})
     layout.layouts = append(layout.layouts, &ViewLayout{View{buffer: b}})
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
                    new_selected_layout := layout.Find(Point{view_selected_layout.view.buffer.cursor.x, view_selected_layout.view.rect.bottom + 2})
                    if new_selected_layout != nil {
                         selected_layout = new_selected_layout
                    }
               case termbox.KeyCtrlK:
                    new_selected_layout := layout.Find(Point{view_selected_layout.view.buffer.cursor.x, view_selected_layout.view.rect.top - 2})
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
                         view_selected_layout.view.cursor = Point{0, len(b.lines) - 1}
                         view_selected_layout.view.cursor = b.ClampOn(view_selected_layout.view.cursor)
                    case '$':
                         view_selected_layout.view.cursor = Point{len(b.lines[b.cursor.y]) - 1, b.cursor.y}
                         view_selected_layout.view.cursor = b.ClampOn(view_selected_layout.view.cursor)
                    case '0':
                         view_selected_layout.view.cursor = Point{0, b.cursor.y}
                         view_selected_layout.view.cursor = b.ClampOn(view_selected_layout.view.cursor)
                    }
               }

               view_selected_layout.view.ScrollTo(view_selected_layout.view.cursor)
          }
     }
}
