package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"go/scanner"
	"go/token"
	"go/types"
	"unicode/utf8"
)

type Highlighter interface {
	Buffer
	Highlight(point Point) (fgColor termbox.Attribute, bgColor termbox.Attribute)
}

type termColor struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

// internal type which wraps a buffer with go syntax highlighting
type goSyntaxHighlighter struct {
	Buffer
	Colors [][]termColor
}

func (syntax *goSyntaxHighlighter) highlightRange(lit string, start token.Position, end token.Position, color termColor) {
	for row := start.Line; row <= end.Line; row++ {
		var startColumn int

		if row == start.Line {
			startColumn = start.Column
		} else {
			startColumn = 1
		}

		var runeIx int
		for x := range syntax.Buffer.Lines()[row-1] {
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

// add go syntax highlighting to the provided buffer
func NewGoSyntaxHighlighter(buffer Buffer) Highlighter {
	syntax := &goSyntaxHighlighter{Buffer: buffer}
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

func (syntax *goSyntaxHighlighter) Highlight(point Point) (termbox.Attribute, termbox.Attribute) {
	return syntax.Colors[point.y][point.x].fg, syntax.Colors[point.y][point.x].bg
}
