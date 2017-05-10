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
