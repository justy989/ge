package edit

type View struct {
	Rect   Rect
	Scroll Point
	Buffer Buffer
	Cursor Point
}

func (view *View) ScrollTo(point Point) {
	view_dimensions := view.Rect.Dimensions()

	if point.Y < view.Scroll.Y {
		view.Scroll.Y = point.Y
	} else {
		// TODO: bobby would like to understand why this works
		view_dimensions.Y--
		bottom := view.Scroll.Y + view_dimensions.Y

		if point.Y > bottom {
			view.Scroll.Y = point.Y - view_dimensions.Y
		}
	}

	if point.X < view.Scroll.X {
		view.Scroll.X = point.X
	} else {
		// TODO: bobby would like to understand why this works
		view_dimensions.X--
		right := view.Scroll.X + view_dimensions.X

		if point.X > right {
			view.Scroll.X = point.X - view_dimensions.X
		}
	}
}
