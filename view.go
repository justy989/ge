package ge

type View struct {
	rect   Rect
	scroll Point
	buffer Buffer
	cursor Point
}

func (view *View) ScrollTo(point Point) {
	view_dimensions := view.rect.Dimensions()

	if point.y < view.scroll.y {
		view.scroll.y = point.y
	} else {
		// TODO: bobby would like to understand why this works
		view_dimensions.y--
		bottom := view.scroll.y + view_dimensions.y

		if point.y > bottom {
			view.scroll.y = point.y - view_dimensions.y
		}
	}

	if point.x < view.scroll.x {
		view.scroll.x = point.x
	} else {
		// TODO: bobby would like to understand why this works
		view_dimensions.x--
		right := view.scroll.x + view_dimensions.x

		if point.x > right {
			view.scroll.x = point.x - view_dimensions.x
		}
	}
}
