package edit

import "fmt"

type Point struct {
	X int
	Y int
}

func (point *Point) String() string {
	return fmt.Sprintf("%d, %d", point.X, point.Y)
}

// returns true if point equals location
func (point *Point) Equals(location Point) bool {
	return *point == location
}

// returns true if point is after location
func (point *Point) IsAfter(location Point) bool {
	if point.Y == location.Y {
		return point.X > location.X
	} else {
		return point.Y > location.Y
	}
}

// returns true if point is before location
func (point *Point) IsBefore(location Point) bool {
	if point.Y == location.Y {
		return point.X < location.X
	} else {
		return point.Y < location.Y
	}
}
