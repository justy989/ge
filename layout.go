package main

import "github.com/nsf/termbox-go"
//import "log"

type Layout interface {
     Rect()(Rect)
     Draw(terminal_dimensions Point)
     CalculateRect(rect Rect)
     Find(query Point) (Layout)
}

type VerticalLayout struct {
	rect    Rect
	layouts []Layout
}

type ViewLayout struct {
	view View
}

type TabLayout struct {
     rect Rect
     root Layout
	selection Layout
     next Layout
     prev Layout
}

func (layout *VerticalLayout) Rect() (Rect) {
     return layout.rect
}

func (layout *VerticalLayout) Draw(terminal_dimensions Point) {
	rect_width := layout.rect.Width()
	for _, child := range layout.layouts {
		child.Draw(terminal_dimensions)
		for i := 0; i < rect_width; i++ {
			termbox.SetCell(layout.rect.left+i, child.Rect().bottom, '-', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func (layout *VerticalLayout) CalculateRect(rect Rect) {
	layout.rect = rect
	sliced_view := rect
	separator_lines := len(layout.layouts) - 1
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
}

func (layout *VerticalLayout) Find(query Point) Layout {
	for _, child := range layout.layouts {
		matched_layout := child.Find(query)
		if matched_layout != nil {
			return matched_layout
		}
	}

	return nil
}

func (layout *ViewLayout) Rect() (Rect) {
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

func (layout *ViewLayout) Find(query Point) Layout {
	if layout.view.rect.Contains(query) {
		return layout
	}

	return nil
}

func (layout *TabLayout) Rect() (Rect) {
     return layout.rect
}

func (layout *TabLayout) Draw(terminal_dimensions Point) {
     // TODO: draw tab bar if we have other tabs
     layout.root.Draw(terminal_dimensions)

     // debuging drawing selection
     rect := layout.selection.Rect()
     termbox.SetCell(rect.left, rect.top, '*', termbox.ColorDefault, termbox.ColorDefault)
     termbox.SetCell(rect.right - 1, rect.top, '*', termbox.ColorDefault, termbox.ColorDefault)
     termbox.SetCell(rect.left, rect.bottom - 1, '*', termbox.ColorDefault, termbox.ColorDefault)
     termbox.SetCell(rect.right - 1, rect.bottom - 1, '*', termbox.ColorDefault, termbox.ColorDefault)
}

func (layout *TabLayout) CalculateRect(rect Rect) {
	layout.rect = rect
     layout.root.CalculateRect(rect)
}

func (layout *TabLayout) Find(query Point) Layout {
     return layout.root.Find(query)
}

func FindViewLayout(itr Layout) (*ViewLayout){
     switch current_node := itr.(type) {
     default:
          panic("unexpected type")
     case *ViewLayout:
          return current_node
     case *VerticalLayout:
          for _, child := range current_node.layouts {
               view_child := FindViewLayout(child)
               if view_child != nil {
                    return view_child
               }
          }
     }

     return nil
}

func SplitLayout(itr Layout, match Layout) {
     switch current_node := itr.(type) {
     default:
          panic("unexpected type")
     case *ViewLayout:
          return
     case *VerticalLayout:
          for i, child := range current_node.layouts {
               if child == match {
                    switch current_child := child.(type) {
                    default:
                         panic("unexpected type");
                    case *ViewLayout:
                         // replace child with a vertical layout that has the child split twice
                         new_layout := &VerticalLayout{}
                         new_layout.layouts = append(new_layout.layouts, &ViewLayout{current_child.view})
                         new_layout.layouts = append(new_layout.layouts, &ViewLayout{current_child.view})
                         current_node.layouts[i] = new_layout
                    case *VerticalLayout:
                         existing_view_layout := FindViewLayout(current_child)
                         if existing_view_layout != nil {
                              current_child.layouts = append(current_child.layouts, &ViewLayout{existing_view_layout.view})
                         } else {
                              panic("no existing view")
                         }
                    }
               } else {
                    SplitLayout(child, match)
               }
          }
     }
}

func (layout *TabLayout) Split() {
     loc := Point{layout.selection.Rect().left, layout.selection.Rect().top}

     if layout.selection == layout {
          existing_view_layout := FindViewLayout(layout.root)
          if existing_view_layout != nil {
               new_layout := &VerticalLayout{}
               new_layout.layouts = append(new_layout.layouts, layout.root)
               new_layout.layouts = append(new_layout.layouts, &ViewLayout{existing_view_layout.view})
               layout.root = new_layout
          } else {
               panic("no existing view")
          }
     } else if layout.selection == layout.root {
          switch current_node := layout.root.(type) {
          default:
               panic("unxpected type");
          case *ViewLayout:
               // replace child with a vertical layout that has the child split twice
               new_layout := &VerticalLayout{}
               new_layout.layouts = append(new_layout.layouts, &ViewLayout{current_node.view})
               new_layout.layouts = append(new_layout.layouts, &ViewLayout{current_node.view})
               layout.root = new_layout
          case *VerticalLayout:
               existing_view_layout := FindViewLayout(current_node)
               if existing_view_layout != nil {
                    current_node.layouts = append(current_node.layouts, &ViewLayout{existing_view_layout.view})
               } else {
                    panic("no existing view")
               }
          }
     } else {
          SplitLayout(layout.root, layout.selection)
     }

     layout.CalculateRect(layout.rect)
     layout.selection = layout.Find(loc)
}

func RemoveLayoutNode (root Layout, itr Layout, match Layout) {
     switch current_node := itr.(type) {
     default:
          panic("unexpected type")
     case *ViewLayout:
          return
     case *VerticalLayout:
          for i, child := range current_node.layouts {
               if child == match {
                    if len(current_node.layouts) > 1 {
                         // remove the selection itr
                         current_node.layouts = append(current_node.layouts[:i], current_node.layouts[i + 1:]...)
                         // TODO: if only 1 node left, make sure to collapse it to a ViewLayout
                    } else {
                         // we're going to remove the last layout, so just remove this vertical layout
                         RemoveLayoutNode(root, root, current_node)
                    }
               } else {
                    RemoveLayoutNode(root, child, match)
               }
          }
     }
}

func (layout *TabLayout) Remove() {
     loc := Point{layout.selection.Rect().left, layout.selection.Rect().top}
     RemoveLayoutNode(layout.root, layout.root, layout.selection)
     layout.CalculateRect(layout.rect)
     layout.selection = layout.Find(loc)
}

func FindLayoutParent(itr Layout, match Layout) (Layout) {
     switch current_node := itr.(type) {
     default:
          panic("unexpected type")
     case *ViewLayout:
          return nil
     case *VerticalLayout:
          for _, child := range current_node.layouts {
               if child == match {
                    return itr
               } else {
                    matched_child := FindLayoutParent(child, match)
                    if matched_child != nil {
                         return matched_child
                    }
               }
          }
     }

     return nil
}

func (layout *TabLayout) Move(direction Direction) {
     switch direction {
     default:
          panic("unexpected direction")
     case DIRECTION_LEFT:
     case DIRECTION_UP:
          new_y := layout.selection.Rect().top - 2
          if new_y < 0 {
               new_y += layout.rect.Height()
          }
          switch current_layout := layout.selection.(type) {
          default:
               panic("unexpected type");
          case *ViewLayout:
               layout.selection = layout.Find(Point{current_layout.view.buffer.Cursor().x, new_y})
          case *VerticalLayout:
               layout.selection = layout.Find(Point{current_layout.Rect().left, new_y})
          }
     case DIRECTION_RIGHT:
     case DIRECTION_DOWN:
          new_y := layout.selection.Rect().bottom + 2
          if new_y > layout.rect.Height() {
               new_y -= layout.rect.Height()
          }
          switch current_layout := layout.selection.(type) {
          default:
               panic("unexpected type");
          case *ViewLayout:
               layout.selection = layout.Find(Point{current_layout.view.buffer.Cursor().x, new_y})
          case *VerticalLayout:
               layout.selection = layout.Find(Point{current_layout.Rect().left, new_y})
          }
     case DIRECTION_IN:
          // find children
          switch current_layout := layout.selection.(type) {
          default:
          case *VerticalLayout:
               layout.selection = current_layout.layouts[0]
          }
     case DIRECTION_OUT:
          // find parent
          if layout.selection == layout.root {
               layout.selection = layout
          } else {
               parent := FindLayoutParent(layout.root, layout.selection)
               if parent != nil {
                    layout.selection = parent 
               }
          }
     }
}

func (layout *TabLayout) AppendTab() {

}
