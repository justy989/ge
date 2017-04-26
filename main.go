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

type Buffer struct {
     lines      []string
     cursor     Point
     scroll     int // TODO: move to view
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

func (buffer* Buffer) Draw(width int, height int) {
     last_line := buffer.scroll + height
     if last_line > len(buffer.lines) {
          last_line = len(buffer.lines)
     }

     for y, line := range buffer.lines[buffer.scroll:last_line] {
          if y >= height {
               break
          }

          for x, ch := range line {
               if x >= width {
                    break
               }

               termbox.SetCell(x, y, ch, termbox.ColorDefault, termbox.ColorDefault)
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

func ScrollTo(scroll int, point Point, view_height int) (int) {
     if point.y < scroll {
          scroll = point.y
     } else {
          // TODO: bobby would like to understand why this works
          view_height--
          bottom := scroll + view_height

          if bottom < point.y {
               scroll = point.y - view_height
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

loop:
     for {
          terminal_width, terminal_height := termbox.Size()
          termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
          b.Draw(terminal_width, terminal_height)
          termbox.SetCursor(b.cursor.x, b.cursor.y - b.scroll)
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

               b.scroll = ScrollTo(b.scroll, b.cursor, terminal_height)
          }
     }
}
