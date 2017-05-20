package main

import (
	"github.com/nsf/termbox-go"
	"go/ast"
	"go/parser"
	"go/token"
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

func (syntax *GoSyntax) highlightRange(start token.Position, end token.Position, fgColor termbox.Attribute, bgColor termbox.Attribute) {
	for row := start.Line; row <= end.Line; row++ {
		curLine := syntax.buffer.Lines()[row-1]
		var startColumn, endColumn int
		if row == start.Line {
			startColumn = start.Column
		} else {
			startColumn = 1
		}
		if row == end.Line {
			endColumn = end.Column
		} else {
			endColumn = len(curLine)
		}

		for i := startColumn; i < endColumn; i++ {
			syntax.fgColors[row-1][i-1] = fgColor
			syntax.bgColors[row-1][i-1] = bgColor
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
	// TODO: send the bytes from our buffer into ParseFile
	f, err := parser.ParseFile(fset, "buffer.go", nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
		return syntax
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		fgColor := termbox.ColorDefault
		bgColor := termbox.ColorDefault
		start := fset.Position(n.Pos())
		end := fset.Position(n.End())
		switch x := n.(type) {
		case *ast.BasicLit:
			switch x.Kind {
			case token.INT:
				// print purple
				fgColor = termbox.ColorMagenta
			case token.FLOAT:
				fgColor = termbox.ColorMagenta
			case token.IMAG:
				fgColor = termbox.ColorMagenta
			case token.CHAR:
				fgColor = termbox.ColorRed
			case token.STRING:
				// print red
				fgColor = termbox.ColorRed
			}
		case *ast.Ident:
			if x.Obj != nil {
				switch x.Obj.Kind {
				case ast.Typ:
				case ast.Fun:
				}
			}
		case *ast.Comment:
			fgColor = termbox.ColorGreen
		case *ast.GenDecl:
			start = fset.Position(x.TokPos)
			end = fset.Position(x.TokPos)
			end.Column += len(x.Tok.String())
			fgColor = termbox.ColorBlue
		case *ast.IfStmt:
			fgColor = termbox.ColorBlue
		case *ast.CaseClause:
			fgColor = termbox.ColorBlue
		default:
			return true
		}
		syntax.highlightRange(start, end, fgColor, bgColor)
		return true
	})

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
