package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"go/scanner"
	"go/token"
	"go/types"
	//"log"
)

func printLen(toPrint rune, settings *DrawSettings) int {
	switch toPrint {
	case '	':
		return settings.tabWidth
	default:
		return 1
	}
}

// TODO: change name of this function
func ConvertX(line string, x int, settings *DrawSettings) int {
	var printCursor int
	for _, ch := range line {
		if x <= 0 {
			break
		}

		printCursor += printLen(ch, settings)

		x--
	}
	return printCursor
}

func PrintableCursor(buffer Buffer, point Point, settings *DrawSettings) Point {
	return Point{x: ConvertX(buffer.Lines()[point.y], point.x, settings), y: point.y}
}

type Highlighter interface {
	Highlight(point Point) (fgColor termbox.Attribute, bgColor termbox.Attribute)
}

type GoSyntax struct {
	buffer   Buffer
	fgColors [][]termbox.Attribute
	bgColors [][]termbox.Attribute
}

func (syntax *GoSyntax) highlightRange(lit string, start token.Position, end token.Position, fgColor termbox.Attribute, bgColor termbox.Attribute) {
	for row := start.Line; row <= end.Line; row++ {
		var startColumn int

		if row == start.Line {
			startColumn = start.Column
		} else {
			startColumn = 1
		}

		var runeIx int
		for x := range syntax.buffer.Lines()[row-1] {
			if startColumn-1 > x {
				// we haven't starting highlighting yet
				runeIx++
				continue
			}

			if row == end.Line && end.Column-1 == x {
				// we are done highlighting
				break
			}

			syntax.fgColors[row-1][runeIx] = fgColor
			syntax.bgColors[row-1][runeIx] = bgColor

			runeIx++
		}
	}
}

func NewHighlighter(buffer Buffer) Highlighter {
	syntax := &GoSyntax{buffer: buffer}
	syntax.fgColors = make([][]termbox.Attribute, len(buffer.Lines()))
	syntax.bgColors = make([][]termbox.Attribute, len(buffer.Lines()))
	for y, lineBytes := range buffer.Lines() {
		line := []rune(lineBytes)
		syntax.fgColors[y] = make([]termbox.Attribute, len(line))
		syntax.bgColors[y] = make([]termbox.Attribute, len(line))

		for x := range line {
			syntax.fgColors[y][x] = termbox.ColorDefault
			syntax.bgColors[y][x] = termbox.ColorDefault
		}
	}

	fset := token.NewFileSet() // positions are relative to fset
	src := buffer.(fmt.Stringer).String()
	f := fset.AddFile("", fset.Base(), len(src))
	var s scanner.Scanner
	s.Init(f, []byte(src), nil /* no error handler. TODO: implement one! */, scanner.ScanComments)
	isGoType := func(t string) bool {
		for _, basic := range types.Typ {
			if t == basic.Name() {
				return true
			}
		}
		return false
	}

	// Repeated calls to Scan yield the token sequence found in the input.
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if pos == token.NoPos {
			panic("aghh")
		}
		fgColor := termbox.ColorDefault
		bgColor := termbox.ColorDefault
		start := fset.Position(pos)
		end := start
		end.Column += len(lit)

		switch {
		case tok == token.COMMENT:
			fgColor = termbox.ColorGreen
		case tok == token.STRING:
			fallthrough
		case tok == token.CHAR:
			fgColor = termbox.ColorRed
		case tok == token.IMAG:
			fallthrough
		case tok == token.FLOAT:
			fallthrough
		case tok == token.INT:
			fgColor = termbox.ColorMagenta
		case tok == token.RETURN:
			fgColor = termbox.ColorYellow
		case tok.IsKeyword():
			fgColor = termbox.ColorBlue
		case lit == "nil":
			fgColor = termbox.ColorRed
		case isGoType(lit):
			fgColor = termbox.ColorBlue
		default:
			continue
		}
		syntax.highlightRange(lit, start, end, fgColor, bgColor)
	}

	return syntax
}

func (syntax *GoSyntax) Highlight(point Point) (termbox.Attribute, termbox.Attribute) {
	return syntax.fgColors[point.y][point.x], syntax.bgColors[point.y][point.x]
}

func DrawBuffer(buffer Buffer, view Rect, scroll Point, terminal_dimensions Point, settings *DrawSettings) (err error) {
	last_row := scroll.y + view.Height()
	if last_row > len(buffer.Lines()) {
		last_row = len(buffer.Lines())
	}

	syntax := NewHighlighter(buffer)

	for y, lineBytes := range buffer.Lines()[scroll.y:last_row] {
		if y >= view.Height() {
			break
		}
		final_y := y + view.top
		if final_y >= terminal_dimensions.y {
			break
		}

		line := []rune(lineBytes)

		var lineWidth, printedWidth int
		for column, ch := range line {
			if printedWidth >= view.Width() {
				break
			}

			final_x := printedWidth + view.left
			if final_x >= terminal_dimensions.x {
				break
			}

			lineWidth += printLen(ch, settings)
			if lineWidth > scroll.x {
				fgColor, bgColor := syntax.Highlight(Point{x: column, y: scroll.y + y})
				termbox.SetCell(final_x, final_y, ch, fgColor, bgColor)
				printedWidth = lineWidth - scroll.x
			}
		}
	}
	return
}
