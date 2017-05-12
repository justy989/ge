package main

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

func calc_cursor_on_terminal(cursor Point, scroll Point, view_top_left Point) Point {
	cursor.x = cursor.x - scroll.x + view_top_left.x
	cursor.y = cursor.y - scroll.y + view_top_left.y
	return cursor
}
