package edit

import "github.com/nsf/termbox-go"

type Layout interface {
	Rect() Rect
	Draw(terminal_dimensions Point, settings *DrawSettings)
	CalculateRect(rect Rect)
	FindView(query Point) Layout
}

type ListLayout struct {
	rect       Rect
	layouts    []Layout
	horizontal bool
}

type ViewLayout struct {
	View View
}

type TabLayout struct {
	rect      Rect
	Root      Layout
	Selection Layout
}

type TabListLayout struct {
	rect      Rect
	Tabs      []TabLayout
	Selection int
}

func (layout *ListLayout) Rect() Rect {
	return layout.rect
}

func (layout *ListLayout) Draw(terminal_dimensions Point, settings *DrawSettings) {
	if layout.horizontal {
		rect_width := layout.rect.Width()
		for _, child := range layout.layouts {
			child.Draw(terminal_dimensions, settings)
			for i := 0; i < rect_width; i++ {
				termbox.SetCell(layout.rect.Left+i, child.Rect().Bottom, '─', termbox.ColorDefault, termbox.ColorDefault)
			}
		}
	} else {
		rect_height := layout.rect.Height()
		for _, child := range layout.layouts {
			child.Draw(terminal_dimensions, settings)
			for i := 0; i < rect_height; i++ {
				termbox.SetCell(child.Rect().Right, layout.rect.Top+i, '│', termbox.ColorDefault, termbox.ColorDefault)
			}
		}
	}
}

func (layout *ListLayout) CalculateRect(rect Rect) {
	layout.rect = rect
	sliced_view := rect
	separator_lines := len(layout.layouts) - 1

	if layout.horizontal {
		slice_height := (rect.Height() - separator_lines) / len(layout.layouts)
		leftover_lines := (rect.Height() - separator_lines) % len(layout.layouts)
		sliced_view.Bottom = sliced_view.Top + slice_height

		// split views evenly
		for _, child := range layout.layouts {
			if leftover_lines > 0 {
				leftover_lines--
				sliced_view.Bottom++
			}

			child.CalculateRect(sliced_view)

			// figure out the next child's view
			sliced_view.Top = sliced_view.Bottom + 1 // account for separator
			sliced_view.Bottom = sliced_view.Top + slice_height
		}
	} else {
		slice_width := (rect.Width() - separator_lines) / len(layout.layouts)
		leftover_lines := (rect.Width() - separator_lines) % len(layout.layouts)
		sliced_view.Right = sliced_view.Left + slice_width

		// split views evenly
		for _, child := range layout.layouts {
			if leftover_lines > 0 {
				leftover_lines--
				sliced_view.Right++
			}

			child.CalculateRect(sliced_view)

			// figure out the next child's view
			sliced_view.Left = sliced_view.Right + 1 // account for separator
			sliced_view.Right = sliced_view.Left + slice_width
		}
	}
}

func (layout *ListLayout) FindView(query Point) Layout {
	for _, child := range layout.layouts {
		matched_layout := child.FindView(query)
		if matched_layout != nil {
			return matched_layout
		}
	}

	return nil
}

func (layout *ListLayout) SetHorizontal(value bool) {
	layout.horizontal = value
}

func (layout *ViewLayout) Rect() Rect {
	return layout.View.Rect
}

func (layout *ViewLayout) Draw(terminal_dimensions Point, settings *DrawSettings) {
	if layout.View.Buffer != nil {
		DrawBuffer(layout.View.Buffer, layout.View.Rect, layout.View.Scroll, terminal_dimensions, settings)
	}
}

func (layout *ViewLayout) CalculateRect(rect Rect) {
	layout.View.Rect = rect
}

func (layout *ViewLayout) FindView(query Point) Layout {
	if layout.View.Rect.Contains(query) {
		return layout
	}

	return nil
}

func (layout *TabLayout) Rect() Rect {
	return layout.rect
}

func (layout *TabLayout) Draw(terminal_dimensions Point, settings *DrawSettings) {
	// TODO: draw tab bar if we have other tabs
	layout.Root.Draw(terminal_dimensions, settings)

	// debuging drawing selection
	rect := layout.Selection.Rect()
	_, is_view_layout := layout.Selection.(*ViewLayout)
	if !is_view_layout {
		fg_color := termbox.ColorWhite | termbox.AttrBold

		for i := rect.Left; i < rect.Right; i++ {
			termbox.SetCell(i, rect.Top, '─', fg_color, termbox.ColorDefault)
			termbox.SetCell(i, rect.Bottom-1, '─', fg_color, termbox.ColorDefault)
		}

		for i := rect.Top; i < rect.Bottom; i++ {
			termbox.SetCell(rect.Left, i, '│', fg_color, termbox.ColorDefault)
			termbox.SetCell(rect.Right-1, i, '│', fg_color, termbox.ColorDefault)
		}
	}

	for i := layout.rect.Left; i < layout.rect.Right; i++ {
		termbox.SetCell(i, layout.rect.Bottom-1, '─', termbox.ColorDefault, termbox.ColorDefault)
	}

	termbox.SetCell(rect.Right-5, rect.Bottom, ' ', termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(rect.Right-2, rect.Bottom, ' ', termbox.ColorDefault, termbox.ColorDefault)

	list_layout, is_selection_list_layout := layout.Selection.(*ListLayout)
	if is_selection_list_layout {
		var first_ch rune
		var second_ch rune
		if len(list_layout.layouts) > 1 {
			first_ch = 'A'
		} else {
			first_ch = 'S'
		}

		if list_layout.horizontal {
			second_ch = 'H'
		} else {
			second_ch = 'V'
		}

		termbox.SetCell(rect.Right-4, rect.Bottom, first_ch, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(rect.Right-3, rect.Bottom, second_ch, termbox.ColorDefault, termbox.ColorDefault)
	}

	// post process to connect the lines
	term_width, term_height := termbox.Size()
	cell_buffer := termbox.CellBuffer()
	var left termbox.Cell
	var right termbox.Cell
	var top termbox.Cell
	var bottom termbox.Cell

	// TODO: optimise only looking at view borders
	for y := 0; y < term_height; y++ {
		for x := 0; x < term_width; x++ {
			center := cell_buffer[y*term_width+x]
			left = termbox.Cell{}
			if x > 0 {
				left = cell_buffer[y*term_width+x-1]
			}

			right = termbox.Cell{}
			if x < (term_width - 1) {
				right = cell_buffer[y*term_width+x+1]
			}

			top = termbox.Cell{}
			if y > 0 {
				top = cell_buffer[(y-1)*term_width+x]
			}

			bottom = termbox.Cell{}
			if y < (term_height - 1) {
				bottom = cell_buffer[(y+1)*term_width+x]
			}

			if left.Ch == '─' && top.Ch == '│' {
				termbox.SetCell(x, y, '┘', center.Fg, center.Bg)
			}

			if left.Ch == '─' && bottom.Ch == '│' {
				termbox.SetCell(x, y, '┐', center.Fg, center.Bg)
			}

			if right.Ch == '─' && top.Ch == '│' {
				termbox.SetCell(x, y, '└', center.Fg, center.Bg)
			}

			if right.Ch == '─' && bottom.Ch == '│' {
				termbox.SetCell(x, y, '┌', center.Fg, center.Bg)
			}

			if left.Ch == '─' && right.Ch == '─' && top.Ch == '│' {
				termbox.SetCell(x, y, '┴', center.Fg, center.Bg)
			}

			if left.Ch == '─' && right.Ch == '─' && bottom.Ch == '│' {
				termbox.SetCell(x, y, '┬', center.Fg, center.Bg)
			}

			if top.Ch == '│' && bottom.Ch == '│' && left.Ch == '─' {
				termbox.SetCell(x, y, '┤', center.Fg, center.Bg)
			}

			if top.Ch == '│' && bottom.Ch == '│' && right.Ch == '─' {
				termbox.SetCell(x, y, '├', center.Fg, center.Bg)
			}

			if top.Ch == '│' && bottom.Ch == '│' && right.Ch == '─' && left.Ch == '─' {
				termbox.SetCell(x, y, '┼', center.Fg, center.Bg)
			}
		}
	}
}

func (layout *TabLayout) CalculateRect(rect Rect) {
	layout.rect = rect
	rect.Bottom-- // account for status line always at the bottom
	layout.Root.CalculateRect(rect)
}

func (layout *TabLayout) FindView(query Point) Layout {
	found := layout.Root.FindView(query)
	if found == nil {
		panic("ahhh we didn't find anything")
	}
	return found
}

func (layout *TabLayout) Split() {
	loc := Point{layout.Selection.Rect().Left, layout.Selection.Rect().Top}

	if layout.Selection == layout.Root {
		switch current_node := layout.Root.(type) {
		default:
			panic("unxpected type")
		case *ViewLayout:
			new_layout := ListLayout{}
			new_layout.layouts = append(new_layout.layouts, &ViewLayout{current_node.View})
			new_layout.layouts = append(new_layout.layouts, &ViewLayout{current_node.View})
			layout.Root = &new_layout
		case *ListLayout:
			existing_view_layout := findViewLayout(current_node)
			if existing_view_layout == nil {
				panic("no existing view")
			}
			current_node.layouts = append(current_node.layouts, &ViewLayout{existing_view_layout.View})
		}
	} else {
		splitLayout(layout.Root, layout.Selection)
	}

	layout.CalculateRect(layout.rect)
	layout.Selection = layout.FindView(loc)
}

func (layout *TabLayout) Remove() {
	if layout.Selection != layout.Root && viewLayoutCount(layout.Root) > 1 {
		loc := Point{layout.Selection.Rect().Left, layout.Selection.Rect().Top}
		removeLayoutNode(layout.Root, layout.Root, layout.Selection)
		layout.CalculateRect(layout.rect)
		layout.Selection = layout.FindView(loc)
	}
}

func (layout *TabLayout) Select(direction Direction) {
	list_selection_layout, is_list_selection := layout.Selection.(*ListLayout)
	if is_list_selection {
		if len(list_selection_layout.layouts) == 1 {
			if list_selection_layout == layout.Root {
				layout.Root = list_selection_layout.layouts[0] // "fuhget abaht it" -Garbage Collector, Circa 57 BC
				layout.Selection = layout.Root
			} else {
				// remove the list layout with only 1 element
				parent := findLayoutParent(layout.Root, layout.Selection)
				list_parent := parent.(*ListLayout)
				for i, child := range list_parent.layouts {
					if child == layout {
						list_parent.layouts[i] = list_selection_layout.layouts[0]
						layout.Selection = list_parent.layouts[i]
						break
					}
				}
			}
		}
	}

	switch direction {
	default:
		panic("unexpected direction")
	case DIRECTION_LEFT:
		new_x := layout.Selection.Rect().Left - 2 // account for separator
		// wrap around
		if new_x < 0 {
			new_x += layout.rect.Width()
		}
		switch current_layout := layout.Selection.(type) {
		default:
			panic("unexpected type")
		case *ViewLayout:
			cursor := Calc_cursor_on_terminal(current_layout.View.Buffer.Cursor(), current_layout.View.Scroll,
				Point{current_layout.View.Rect.Left, current_layout.View.Rect.Top})
			layout.Selection = layout.FindView(Point{new_x, cursor.Y})
		case *ListLayout:
			layout.Selection = layout.FindView(Point{new_x, current_layout.Rect().Top})
		}
	case DIRECTION_UP:
		new_y := layout.Selection.Rect().Top - 2 // account for separator
		// wrap around
		if new_y < 0 {
			new_y += layout.rect.Height()
		}
		switch current_layout := layout.Selection.(type) {
		default:
			panic("unexpected type")
		case *ViewLayout:
			cursor := Calc_cursor_on_terminal(current_layout.View.Buffer.Cursor(), current_layout.View.Scroll,
				Point{current_layout.View.Rect.Left, current_layout.View.Rect.Top})
			layout.Selection = layout.FindView(Point{cursor.X, new_y})
		case *ListLayout:
			layout.Selection = layout.FindView(Point{current_layout.Rect().Left, new_y})
		}
	case DIRECTION_RIGHT:
		new_x := layout.Selection.Rect().Right + 2 // account for separator
		// wrap around
		if new_x > layout.rect.Width() {
			new_x -= layout.rect.Width()
		}
		switch current_layout := layout.Selection.(type) {
		default:
			panic("unexpected type")
		case *ViewLayout:
			cursor := Calc_cursor_on_terminal(current_layout.View.Buffer.Cursor(), current_layout.View.Scroll,
				Point{current_layout.View.Rect.Left, current_layout.View.Rect.Top})
			layout.Selection = layout.FindView(Point{new_x, cursor.Y})
		case *ListLayout:
			layout.Selection = layout.FindView(Point{new_x, current_layout.Rect().Top})
		}
	case DIRECTION_DOWN:
		new_y := layout.Selection.Rect().Bottom + 2 // account for separator
		// wrap around
		if new_y > layout.rect.Height() {
			new_y -= layout.rect.Height()
		}
		switch current_layout := layout.Selection.(type) {
		default:
			panic("unexpected type")
		case *ViewLayout:
			cursor := Calc_cursor_on_terminal(current_layout.View.Buffer.Cursor(), current_layout.View.Scroll,
				Point{current_layout.View.Rect.Left, current_layout.View.Rect.Top})
			layout.Selection = layout.FindView(Point{cursor.X, new_y})
		case *ListLayout:
			layout.Selection = layout.FindView(Point{current_layout.Rect().Left, new_y})
		}
	case DIRECTION_IN:
		// find children
		switch current_layout := layout.Selection.(type) {
		default:
		case *ListLayout:
			layout.Selection = current_layout.layouts[0]
		}
	case DIRECTION_OUT:
		// find parent
		parent := findLayoutParent(layout.Root, layout.Selection)
		if parent != nil {
			layout.Selection = parent
		}
	}
}

func (layout *TabLayout) PrepareSplit(horizontal bool) {
	list_layout, is_list_layout := layout.Selection.(*ListLayout)
	if is_list_layout {
		if len(list_layout.layouts) == 1 {
			list_layout.horizontal = horizontal
			return
		}
	}

	new_layout := ListLayout{}
	new_layout.horizontal = horizontal
	new_layout.layouts = append(new_layout.layouts, layout.Selection)
	defer func() { layout.Selection = &new_layout }()

	if layout.Root == layout.Selection {
		layout.Root = &new_layout
		return
	}

	parent := findLayoutParent(layout.Root, layout.Selection)

	switch current_parent := parent.(type) {
	default:
		panic("unexpected type")
	case *ListLayout:
		for i, view_layout := range current_parent.layouts {
			if view_layout == layout.Selection {
				current_parent.layouts[i] = &new_layout
				break
			}
		}
	}
}

func (layout *TabListLayout) Rect() Rect {
	return layout.rect
}

func (layout *TabListLayout) Draw(terminal_dimensions Point, settings *DrawSettings) {
	if len(layout.Tabs) > 1 {
		for i := 0; i < layout.rect.Right; i++ {
			termbox.SetCell(i, layout.rect.Top, '─', termbox.ColorDefault, termbox.ColorDefault)
		}

		x := 2
		termbox.SetCell(x, layout.rect.Top, ' ', termbox.ColorDefault, termbox.ColorDefault)
		x++
		for i := range layout.Tabs {
			if i == layout.Selection {
				termbox.SetCell(x, layout.rect.Top, rune(i)+'1', termbox.ColorCyan, termbox.ColorDefault)
			} else {
				termbox.SetCell(x, layout.rect.Top, rune(i)+'1', termbox.ColorDefault, termbox.ColorDefault)
			}
			x++
			termbox.SetCell(x, layout.rect.Top, ' ', termbox.ColorDefault, termbox.ColorDefault)
			x++
		}
	}

	layout.Tabs[layout.Selection].Draw(terminal_dimensions, settings)
}

func (layout *TabListLayout) CalculateRect(rect Rect) {
	layout.rect = rect

	if len(layout.Tabs) > 1 {
		rect.Top++
	}

	layout.Tabs[layout.Selection].CalculateRect(rect)
}

func (layout *TabListLayout) FindView(query Point) Layout {
	return layout.Tabs[layout.Selection].FindView(query)
}

func findViewLayout(itr Layout) *ViewLayout {
	switch current_node := itr.(type) {
	default:
		panic("unexpected type")
	case *ViewLayout:
		return current_node
	case *ListLayout:
		for _, child := range current_node.layouts {
			view_child := findViewLayout(child)
			if view_child != nil {
				return view_child
			}
		}
	}

	return nil
}

func splitLayout(itr Layout, match Layout) {
	switch current_node := itr.(type) {
	default:
		panic("unexpected type")
	case *ViewLayout:
		return
	case *ListLayout:
		for _, child := range current_node.layouts {
			if child == match {
				switch current_child := child.(type) {
				default:
					panic("unexpected type")
				case *ViewLayout:
					current_node.layouts = append(current_node.layouts, &ViewLayout{current_child.View})
				case *ListLayout:
					existing_view_layout := findViewLayout(current_child)
					if existing_view_layout == nil {
						panic("no existing view")
					}
					current_child.layouts = append(current_child.layouts, &ViewLayout{existing_view_layout.View})
				}
			} else {
				splitLayout(child, match)
			}
		}
	}
}

func findLayoutParent(itr Layout, match Layout) Layout {
	switch current_node := itr.(type) {
	default:
		panic("unexpected type")
	case *ViewLayout:
		return nil
	case *ListLayout:
		for _, child := range current_node.layouts {
			if child == match {
				return itr
			} else {
				matched_child := findLayoutParent(child, match)
				if matched_child != nil {
					return matched_child
				}
			}
		}
	}

	return nil
}

func removeLayoutNode(root Layout, itr Layout, match Layout) {
	switch current_node := itr.(type) {
	default:
		panic("unexpected type")
	case *ViewLayout:
		return
	case *ListLayout:
		for i, child := range current_node.layouts {
			if child == match {
				// remove the selection itr
				current_node.layouts = append(current_node.layouts[:i], current_node.layouts[i+1:]...)

				// collapse current list layout if only one is left
				if len(current_node.layouts) == 0 {
					removeLayoutNode(root, root, current_node)
				}
			} else {
				removeLayoutNode(root, child, match)
			}
		}
	}
}

func viewLayoutCount(itr Layout) int {
	switch current_node := itr.(type) {
	default:
	case *ViewLayout:
		return 1
	case *ListLayout:
		count := 0
		for _, child := range current_node.layouts {
			count += viewLayoutCount(child)
		}
		return count
	}

	return 0
}
