package main

import (
	"github.com/nsf/termbox-go"
)

func printLen(toPrint rune, settings *DrawSettings) int {
	switch toPrint {
	case '	':
		return settings.tabWidth
	default:
		return 1
	}
}

func DrawBuffer(buffer Buffer, view Rect, scroll Point, terminal_dimensions Point, settings *DrawSettings) (err error) {
	last_row := scroll.y + view.Height()
	if last_row > len(buffer.Lines()) {
		last_row = len(buffer.Lines())
	}

	for y, line := range buffer.Lines()[scroll.y:last_row] {
		if y >= view.Height() {
			break
		}
		final_y := y + view.top
		if final_y >= terminal_dimensions.y {
			break
		}

		var lineWidth, printedWidth int
		for _, ch := range line {
			if printedWidth >= view.Width() {
				break
			}

			final_x := printedWidth + view.left
			if final_x >= terminal_dimensions.x {
				break
			}

			lineWidth += printLen(ch, settings)
			if lineWidth > scroll.x {
				termbox.SetCell(final_x, final_y, ch, termbox.ColorDefault, termbox.ColorDefault)
				printedWidth = lineWidth - scroll.x
			}
		}
	}
	return
}
