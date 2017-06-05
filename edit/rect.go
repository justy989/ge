package edit

type Rect struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

func NewRectFromPointAndDimension(point Point, dimensions Point) Rect {
	return Rect{point.X, point.Y, point.X + dimensions.X, point.Y + dimensions.Y}
}

func (rect *Rect) Width() int {
	return rect.Right - rect.Left
}

func (rect *Rect) Height() int {
	return rect.Bottom - rect.Top
}

func (rect *Rect) Dimensions() Point {
	return Point{rect.Width(), rect.Height()}
}

func (rect *Rect) Contains(p Point) bool {
	if p.X >= rect.Left && p.X <= rect.Right &&
		p.Y >= rect.Top && p.Y <= rect.Bottom {
		return true
	}

	return false
}
