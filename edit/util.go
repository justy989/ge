package edit

type Direction int

const (
	DIRECTION_LEFT Direction = iota
	DIRECTION_UP
	DIRECTION_RIGHT
	DIRECTION_DOWN
	DIRECTION_IN
	DIRECTION_OUT
)

func Clamp(value int, min int, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}

	return value
}

func Calc_cursor_on_terminal(cursor Point, scroll Point, view_top_left Point) Point {
	cursor.X = cursor.X - scroll.X + view_top_left.X
	cursor.Y = cursor.Y - scroll.Y + view_top_left.Y
	return cursor
}
