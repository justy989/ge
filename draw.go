package main

import (
	//"fmt"
	"github.com/nsf/termbox-go"
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

func DrawBuffer(buffer Buffer, view Rect, scroll Point, terminal_dimensions Point, settings *DrawSettings) (err error) {
	last_row := scroll.y + view.Height()
	if last_row > len(buffer.Lines()) {
		last_row = len(buffer.Lines())
	}

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
				var fgColor, bgColor termbox.Attribute
				// TODO: figure out how to make wrapper-like interfaces
				undoer, ok := buffer.(*undoBuffer)
				if ok {
					highlighter, ok := undoer.Buffer.(Highlighter)
					if ok {
						fgColor, bgColor = highlighter.Highlight(Point{x: column, y: scroll.y + y})
					}

				}
				termbox.SetCell(final_x, final_y, ch, fgColor, bgColor)
				printedWidth = lineWidth - scroll.x
			}
		}
	}
	return
}
