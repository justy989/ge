package edit

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"go/scanner"
	"go/token"
	"go/types"
	"unicode/utf8"
	//"log"
)

func printLen(toPrint rune, settings *DrawSettings) int {
	switch toPrint {
	case '	':
		return settings.TabWidth
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
	return Point{X: ConvertX(buffer.Lines()[point.Y], point.X, settings), Y: point.Y}
}

type Highlighter interface {
	Highlight(point Point) (fgColor termbox.Attribute, bgColor termbox.Attribute)
}

type termColor struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

type GoSyntax struct {
	buffer Buffer
	Colors [][]termColor
}

func (syntax *GoSyntax) highlightRange(lit string, start token.Position, end token.Position, color termColor) {
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

			syntax.Colors[row-1][runeIx] = color

			runeIx++
		}
	}
}

func NewHighlighter(buffer Buffer) Highlighter {
	syntax := &GoSyntax{buffer: buffer}
	syntax.Colors = make([][]termColor, len(buffer.Lines()))
	for y, lineBytes := range buffer.Lines() {
		lineLen := utf8.RuneCountInString(lineBytes)
		syntax.Colors[y] = make([]termColor, lineLen)
	}

	fset := token.NewFileSet() // positions are relative to fset
	src := buffer.(fmt.Stringer).String()
	f := fset.AddFile("", fset.Base(), len(src))
	var s scanner.Scanner
	s.Init(f, []byte(src), nil /* no error handler. TODO: implement one! */, scanner.ScanComments)
	goTypes := make(map[string]struct{})
	for _, basic := range types.Typ {
		// add empty struct to map
		goTypes[basic.Name()] = struct{}{}
	}

	isGoType := func(t string) bool {
		_, ok := goTypes[t]
		return ok
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
		color := termColor{}
		start := fset.Position(pos)
		end := start
		end.Column += len(lit)

		switch {
		case tok == token.COMMENT:
			color.fg = termbox.ColorGreen
		case tok == token.STRING:
			fallthrough
		case tok == token.CHAR:
			color.fg = termbox.ColorRed
		case tok == token.IMAG:
			fallthrough
		case tok == token.FLOAT:
			fallthrough
		case tok == token.INT:
			color.fg = termbox.ColorMagenta
		case tok == token.RETURN:
			color.fg = termbox.ColorYellow
		case tok.IsKeyword():
			color.fg = termbox.ColorBlue
		case lit == "nil":
			color.fg = termbox.ColorRed
		case isGoType(lit):
			color.fg = termbox.ColorBlue
		default:
			continue
		}
		syntax.highlightRange(lit, start, end, color)
	}

	return syntax
}

func (syntax *GoSyntax) Highlight(point Point) (termbox.Attribute, termbox.Attribute) {
	return syntax.Colors[point.Y][point.X].fg, syntax.Colors[point.Y][point.X].bg
}

func DrawBuffer(buffer Buffer, view Rect, scroll Point, terminal_dimensions Point, settings *DrawSettings) (err error) {
	last_row := scroll.Y + view.Height()
	if last_row > len(buffer.Lines()) {
		last_row = len(buffer.Lines())
	}

	syntax := NewHighlighter(buffer)

	for y, lineBytes := range buffer.Lines()[scroll.Y:last_row] {
		if y >= view.Height() {
			break
		}
		final_y := y + view.Top
		if final_y >= terminal_dimensions.Y {
			break
		}

		line := []rune(lineBytes)

		var lineWidth, printedWidth int
		for column, ch := range line {
			if printedWidth >= view.Width() {
				break
			}

			final_x := printedWidth + view.Left
			if final_x >= terminal_dimensions.X {
				break
			}

			lineWidth += printLen(ch, settings)
			if lineWidth > scroll.X {
				fgColor, bgColor := syntax.Highlight(Point{X: column, Y: scroll.Y + y})
				termbox.SetCell(final_x, final_y, ch, fgColor, bgColor)
				printedWidth = lineWidth - scroll.X
			}
		}
	}
	return
}
