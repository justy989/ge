package main

import (
     "io"
     "bufio"
     "bytes"
)

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

     for scanner.Scan() {
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

func (buffer* Buffer) MoveCursor(cursor Point, delta Point) (Point) {
     final := Point{cursor.x + delta.x, cursor.y + delta.y}
     cursor = buffer.ClampOn(final)
     return cursor
}
