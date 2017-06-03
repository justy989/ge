package ge

import "fmt"

type Point struct {
	x int
	y int
}

func (point *Point) String() string {
	return fmt.Sprintf("%d, %d", point.x, point.y)
}

// returns true if point equals location
func (point *Point) Equals(location Point) bool {
	return *point == location
}

// returns true if point is after location
func (point *Point) IsAfter(location Point) bool {
	if point.y == location.y {
		return point.x > location.x
	} else {
		return point.y > location.y
	}
}

// returns true if point is before location
func (point *Point) IsBefore(location Point) bool {
	if point.y == location.y {
		return point.x < location.x
	} else {
		return point.y < location.y
	}
}
