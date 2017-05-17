package main

import "github.com/nsf/termbox-go"

type Layout interface {
	Rect() Rect
	Draw(terminal_dimensions Point)
	CalculateRect(rect Rect)
	FindView(query Point) Layout
}

type ListLayout struct {
	rect       Rect
	layouts    []Layout
	horizontal bool
}

type ViewLayout struct {
	view View
}

type TabLayout struct {
	rect      Rect
	root      Layout
	selection Layout
}

type TabListLayout struct {
	rect      Rect
	tabs      []TabLayout
	selection int
}

func (layout *ListLayout) Rect() Rect {
	return layout.rect
}

func (layout *ListLayout) Draw(terminal_dimensions Point) {
	if layout.horizontal {
		rect_width := layout.rect.Width()
		for _, child := range layout.layouts {
			child.Draw(terminal_dimensions)
			for i := 0; i < rect_width; i++ {
				termbox.SetCell(layout.rect.left+i, child.Rect().bottom, '─', termbox.ColorDefault, termbox.ColorDefault)
			}
		}
	} else {
		rect_height := layout.rect.Height()
		for _, child := range layout.layouts {
			child.Draw(terminal_dimensions)
			for i := 0; i < rect_height; i++ {
				termbox.SetCell(child.Rect().right, layout.rect.top+i, '│', termbox.ColorDefault, termbox.ColorDefault)
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
		sliced_view.bottom = sliced_view.top + slice_height

		// split views evenly
		for _, child := range layout.layouts {
			if leftover_lines > 0 {
				leftover_lines--
				sliced_view.bottom++
			}

			child.CalculateRect(sliced_view)

			// figure out the next child's view
			sliced_view.top = sliced_view.bottom + 1 // account for separator
			sliced_view.bottom = sliced_view.top + slice_height
		}
	} else {
		slice_width := (rect.Width() - separator_lines) / len(layout.layouts)
		leftover_lines := (rect.Width() - separator_lines) % len(layout.layouts)
		sliced_view.right = sliced_view.left + slice_width

		// split views evenly
		for _, child := range layout.layouts {
			if leftover_lines > 0 {
				leftover_lines--
				sliced_view.right++
			}

			child.CalculateRect(sliced_view)

			// figure out the next child's view
			sliced_view.left = sliced_view.right + 1 // account for separator
			sliced_view.right = sliced_view.left + slice_width
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
	return layout.view.rect
}

func (layout *ViewLayout) Draw(terminal_dimensions Point) {
	if layout.view.buffer != nil {
		layout.view.buffer.Draw(layout.view.rect, layout.view.scroll, terminal_dimensions)
	}
}

func (layout *ViewLayout) CalculateRect(rect Rect) {
	layout.view.rect = rect
}

func (layout *ViewLayout) FindView(query Point) Layout {
	if layout.view.rect.Contains(query) {
		return layout
	}

	return nil
}

func (layout *TabLayout) Rect() Rect {
	return layout.rect
}

func (layout *TabLayout) Draw(terminal_dimensions Point) {
	// TODO: draw tab bar if we have other tabs
	layout.root.Draw(terminal_dimensions)

	// debuging drawing selection
	rect := layout.selection.Rect()
	_, is_view_layout := layout.selection.(*ViewLayout)
	if !is_view_layout {
		fg_color := termbox.ColorWhite | termbox.AttrBold

		for i := rect.left; i < rect.right; i++ {
			termbox.SetCell(i, rect.top, '─', fg_color, termbox.ColorDefault)
			termbox.SetCell(i, rect.bottom-1, '─', fg_color, termbox.ColorDefault)
		}

		for i := rect.top; i < rect.bottom; i++ {
			termbox.SetCell(rect.left, i, '│', fg_color, termbox.ColorDefault)
			termbox.SetCell(rect.right-1, i, '│', fg_color, termbox.ColorDefault)
		}
	}

	for i := layout.rect.left; i < layout.rect.right; i++ {
		termbox.SetCell(i, layout.rect.bottom-1, '─', termbox.ColorDefault, termbox.ColorDefault)
	}

	termbox.SetCell(rect.right-5, rect.bottom, ' ', termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(rect.right-2, rect.bottom, ' ', termbox.ColorDefault, termbox.ColorDefault)

	list_layout, is_selection_list_layout := layout.selection.(*ListLayout)
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

		termbox.SetCell(rect.right-4, rect.bottom, first_ch, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(rect.right-3, rect.bottom, second_ch, termbox.ColorDefault, termbox.ColorDefault)
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
	rect.bottom-- // account for status line always at the bottom
	layout.root.CalculateRect(rect)
}

func (layout *TabLayout) FindView(query Point) Layout {
	found := layout.root.FindView(query)
	if found == nil {
		panic("ahhh we didn't find anything")
	}
	return found
}

func (layout *TabLayout) Split() {
	loc := Point{layout.selection.Rect().left, layout.selection.Rect().top}

	if layout.selection == layout.root {
		switch current_node := layout.root.(type) {
		default:
			panic("unxpected type")
		case *ViewLayout:
			new_layout := ListLayout{}
			new_layout.layouts = append(new_layout.layouts, &ViewLayout{current_node.view})
			new_layout.layouts = append(new_layout.layouts, &ViewLayout{current_node.view})
			layout.root = &new_layout
		case *ListLayout:
			existing_view_layout := findViewLayout(current_node)
			if existing_view_layout == nil {
				panic("no existing view")
			}
			current_node.layouts = append(current_node.layouts, &ViewLayout{existing_view_layout.view})
		}
	} else {
		splitLayout(layout.root, layout.selection)
	}

	layout.CalculateRect(layout.rect)
	layout.selection = layout.FindView(loc)
}

func (layout *TabLayout) Remove() {
	if layout.selection != layout.root && viewLayoutCount(layout.root) > 1 {
		loc := Point{layout.selection.Rect().left, layout.selection.Rect().top}
		removeLayoutNode(layout.root, layout.root, layout.selection)
		layout.CalculateRect(layout.rect)
		layout.selection = layout.FindView(loc)
	}
}

func (layout *TabLayout) Select(direction Direction) {
	switch direction {
	default:
		panic("unexpected direction")
	case DIRECTION_LEFT:
		new_x := layout.selection.Rect().left - 2 // account for separator
		// wrap around
		if new_x < 0 {
			new_x += layout.rect.Width()
		}
		switch current_layout := layout.selection.(type) {
		default:
			panic("unexpected type")
		case *ViewLayout:
			cursor := calc_cursor_on_terminal(current_layout.view.buffer.Cursor(), current_layout.view.scroll,
				Point{current_layout.view.rect.left, current_layout.view.rect.top})
			layout.selection = layout.FindView(Point{new_x, cursor.y})
		case *ListLayout:
			if len(current_layout.layouts) == 1 {
				layout.Remove()
			}

			layout.selection = layout.FindView(Point{new_x, current_layout.Rect().top})
		}
	case DIRECTION_UP:
		new_y := layout.selection.Rect().top - 2 // account for separator
		// wrap around
		if new_y < 0 {
			new_y += layout.rect.Height()
		}
		switch current_layout := layout.selection.(type) {
		default:
			panic("unexpected type")
		case *ViewLayout:
			cursor := calc_cursor_on_terminal(current_layout.view.buffer.Cursor(), current_layout.view.scroll,
				Point{current_layout.view.rect.left, current_layout.view.rect.top})
			layout.selection = layout.FindView(Point{cursor.x, new_y})
		case *ListLayout:
			layout.selection = layout.FindView(Point{current_layout.Rect().left, new_y})
		}
	case DIRECTION_RIGHT:
		new_x := layout.selection.Rect().right + 2 // account for separator
		// wrap around
		if new_x > layout.rect.Width() {
			new_x -= layout.rect.Width()
		}
		switch current_layout := layout.selection.(type) {
		default:
			panic("unexpected type")
		case *ViewLayout:
			cursor := calc_cursor_on_terminal(current_layout.view.buffer.Cursor(), current_layout.view.scroll,
				Point{current_layout.view.rect.left, current_layout.view.rect.top})
			layout.selection = layout.FindView(Point{new_x, cursor.y})
		case *ListLayout:
			layout.selection = layout.FindView(Point{new_x, current_layout.Rect().top})
		}
	case DIRECTION_DOWN:
		new_y := layout.selection.Rect().bottom + 2 // account for separator
		// wrap around
		if new_y > layout.rect.Height() {
			new_y -= layout.rect.Height()
		}
		switch current_layout := layout.selection.(type) {
		default:
			panic("unexpected type")
		case *ViewLayout:
			cursor := calc_cursor_on_terminal(current_layout.view.buffer.Cursor(), current_layout.view.scroll,
				Point{current_layout.view.rect.left, current_layout.view.rect.top})
			layout.selection = layout.FindView(Point{cursor.x, new_y})
		case *ListLayout:
			layout.selection = layout.FindView(Point{current_layout.Rect().left, new_y})
		}
	case DIRECTION_IN:
		// find children
		switch current_layout := layout.selection.(type) {
		default:
		case *ListLayout:
			layout.selection = current_layout.layouts[0]
		}
	case DIRECTION_OUT:
		// find parent
		parent := findLayoutParent(layout.root, layout.selection)
		if parent != nil {
			layout.selection = parent
		}
	}
}

func (layout *TabLayout) PrepareSplit(horizontal bool) {
	list_layout, is_list_layout := layout.selection.(*ListLayout)
	if is_list_layout {
		if len(list_layout.layouts) == 1 {
			list_layout.horizontal = horizontal
			return
		}
	}

	new_layout := ListLayout{}
	new_layout.horizontal = horizontal
	new_layout.layouts = append(new_layout.layouts, layout.selection)
	defer func() { layout.selection = &new_layout }()

	if layout.root == layout.selection {
		layout.root = &new_layout
		return
	}

	parent := findLayoutParent(layout.root, layout.selection)

	switch current_parent := parent.(type) {
	default:
		panic("unexpected type")
	case *ListLayout:
		for i, view_layout := range current_parent.layouts {
			if view_layout == layout.selection {
				current_parent.layouts[i] = &new_layout
				break
			}
		}
	}
}

func (layout *TabListLayout) Rect() Rect {
	return layout.rect
}

func (layout *TabListLayout) Draw(terminal_dimensions Point) {
	if len(layout.tabs) > 1 {
		for i := 0; i < layout.rect.right; i++ {
			termbox.SetCell(i, layout.rect.top, '─', termbox.ColorDefault, termbox.ColorDefault)
		}

		x := 2
		termbox.SetCell(x, layout.rect.top, ' ', termbox.ColorDefault, termbox.ColorDefault)
		x++
		for i := range layout.tabs {
			if i == layout.selection {
				termbox.SetCell(x, layout.rect.top, rune(i)+'1', termbox.ColorCyan, termbox.ColorDefault)
			} else {
				termbox.SetCell(x, layout.rect.top, rune(i)+'1', termbox.ColorDefault, termbox.ColorDefault)
			}
			x++
			termbox.SetCell(x, layout.rect.top, ' ', termbox.ColorDefault, termbox.ColorDefault)
			x++
		}
	}

	layout.tabs[layout.selection].Draw(terminal_dimensions)
}

func (layout *TabListLayout) CalculateRect(rect Rect) {
	layout.rect = rect

	if len(layout.tabs) > 1 {
		rect.top++
	}

	layout.tabs[layout.selection].CalculateRect(rect)
}

func (layout *TabListLayout) FindView(query Point) Layout {
	return layout.tabs[layout.selection].FindView(query)
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
					current_node.layouts = append(current_node.layouts, &ViewLayout{current_child.view})
				case *ListLayout:
					existing_view_layout := findViewLayout(current_child)
					if existing_view_layout == nil {
						panic("no existing view")
					}
					current_child.layouts = append(current_child.layouts, &ViewLayout{existing_view_layout.view})
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
