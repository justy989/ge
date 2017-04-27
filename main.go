package main

import (
     "fmt"
     "io"
     "bufio"
     "os"
     "bytes"
     "github.com/nsf/termbox-go"
)

type Point struct {
     x int
     y int
}

func (point *Point) String() (string) {
     return fmt.Sprintf("%d, %d", point.x, point.y)
}

type Rect struct {
     left int
     top int
     right int
     bottom int
}

func NewRectFromPointAndDimension(point Point, dimensions Point) (Rect){
     return Rect{point.x, point.y, point.x + dimensions.x, point.y + dimensions.y}
}

func (rect* Rect) Width() (int){
     return rect.right - rect.left
}

func (rect* Rect) Height() (int){
     return rect.bottom - rect.top
}

func (rect* Rect) Dimensions() (Point){
     return Point{rect.Width(), rect.Height()}
}

type Buffer struct {
     lines      []string
     cursor     Point
}

func (buffer *Buffer) Write(bytes []byte) (int, error) {
     buffer.lines = append(buffer.lines, string(bytes))
     return len(bytes), nil
}

func NewFileBuffer(r io.Reader) (buffer *Buffer) {
     buffer = &Buffer{}
     scanner := bufio.NewScanner(r)

     for scanner.Scan(){
          buffer.Write(scanner.Bytes())
     }

     return buffer
}

func (buffer *Buffer) String() (string) {
     var b bytes.Buffer
     for _, line := range buffer.lines {
          b.WriteString(line);
     }
     return b.String()
}

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

func Clamp(value int, min int, max int) (int) {
     if value < min {
          return min
     } else if value > max {
          return max
     }

     return value
}

func (buffer *Buffer) ClampOn(point Point) (p Point) {
     p.y = Clamp(point.y, 0, len(buffer.lines) - 1)
     p.x = Clamp(point.x, 0, len(buffer.lines[p.y]))
     return
}

func (buffer *Buffer) ClampIn(point Point) (p Point) {
     p.y = Clamp(point.y, 0, len(buffer.lines) - 1)
     p.x = Clamp(point.x, 0, len(buffer.lines[p.y]) - 1)
     return
}

func (buffer* Buffer) MoveCursor(delta Point) {
     final := Point{buffer.cursor.x + delta.x, buffer.cursor.y + delta.y}
     buffer.cursor = buffer.ClampOn(final)
}

func (buffer* Buffer) SetCursor(point Point) {
     buffer.cursor = buffer.ClampOn(point)
}

func ScrollTo(point Point, scroll Point, view_dimensions Point) (Point) {
     if point.y < scroll.y {
          scroll.y = point.y
     } else {
          // TODO: bobby would like to understand why this works
          view_dimensions.y--
          bottom := scroll.y + view_dimensions.y

          if point.y > bottom {
               scroll.y = point.y - view_dimensions.y
          }
     }

     if point.x < scroll.x {
          scroll.x = point.x
     } else {
          // TODO: bobby would like to understand why this works
          view_dimensions.x--
          right := scroll.x + view_dimensions.x

          if point.x > right {
               scroll.x = point.x - view_dimensions.x
          }
     }

     return scroll
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

     // TODO: make view
     view := Rect{0, 0, terminal_dimensions.x, terminal_dimensions.y}
     scroll := Point{0, 0}

loop:
     for {
          terminal_dimensions.x, terminal_dimensions.y = termbox.Size()
          view.right = terminal_dimensions.x
          view.bottom = terminal_dimensions.y
          termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
          b.Draw(view, scroll, terminal_dimensions)
          termbox.SetCursor(b.cursor.x - scroll.x + view.left, b.cursor.y - scroll.y + view.top)
          termbox.Flush()

          switch ev := termbox.PollEvent(); ev.Type {
          case termbox.EventKey:
               switch ev.Key {
               case termbox.KeyEsc:
                    break loop
               default:
                    switch ev.Ch {
                    case 'h':
                         b.MoveCursor(Point{-1, 0})
                    case 'l':
                         b.MoveCursor(Point{1, 0})
                    case 'k':
                         b.MoveCursor(Point{0, -1})
                    case 'j':
                         b.MoveCursor(Point{0, 1})
                    case 'G':
                         b.SetCursor(Point{0, len(b.lines) - 1})
                    case '$':
                         b.SetCursor(Point{len(b.lines[b.cursor.y]) - 1, b.cursor.y})
                    case '0':
                         b.SetCursor(Point{0, b.cursor.y})
                    }
               }

               scroll = ScrollTo(b.cursor, scroll, view.Dimensions())
          }
     }
}
