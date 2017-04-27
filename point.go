package main

import "fmt"

type Point struct {
     x int
     y int
}

func (point *Point) String() (string) {
     return fmt.Sprintf("%d, %d", point.x, point.y)
}
