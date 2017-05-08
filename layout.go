package main

import "github.com/nsf/termbox-go"
//import "log"

type Layout interface {
     View()(*View)
     Draw(terminal_dimensions Point)
     CalculateView(view Rect)
     Find(query Point) (Layout)
     GetParentOf(layout Layout) (Layout)
}

type Splitter interface {
     Split(layout Layout) (Layout)
}

type VerticalLayout struct {
	view    View
	layouts []Layout
}

type ViewLayout struct {
	view View
}

type MainLayout struct {
	Layout
	selected Layout
}

func (layout *VerticalLayout) View() (view *View) {
	return &layout.view
}

func (layout *VerticalLayout) Draw(terminal_dimensions Point) {
	view_width := layout.view.rect.Width()
	for _, child := range layout.layouts {
		child.Draw(terminal_dimensions)
		for i := 0; i < view_width; i++ {
			termbox.SetCell(layout.view.rect.left+i, child.View().rect.bottom, '-', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func (layout *VerticalLayout) CalculateView(view Rect) {
	layout.view.rect = view
	sliced_view := view
	separator_lines := len(layout.layouts) - 1
	slice_height := (view.Height() - separator_lines) / len(layout.layouts)
	leftover_lines := (view.Height() - separator_lines) % len(layout.layouts)
	sliced_view.bottom = sliced_view.top + slice_height

	// split views evenly
	for _, child := range layout.layouts {
		if leftover_lines > 0 {
			leftover_lines--
			sliced_view.bottom++
		}

		child.CalculateView(sliced_view)
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

func (layout *VerticalLayout) Split(to_split Layout) (Layout){
     if layout == to_split {
          layout.layouts = append(layout.layouts, &ViewLayout{view : layout.view})
          return layout.layouts[len(layout.layouts) - 1]
     }

     for i, child := range layout.layouts {
          if child == to_split {
               splitter, ok := child.(Splitter)
               if !ok {
                    // assuming vertical layout for now
                    replacement_layout := &VerticalLayout{}
                    replacement_layout.layouts = append(replacement_layout.layouts, &ViewLayout{*child.View()})
                    replacement_layout.layouts = append(replacement_layout.layouts, &ViewLayout{*child.View()})
                    layout.layouts[i] = replacement_layout
                    return replacement_layout.layouts[0]
               }
               return splitter.Split(to_split)
          } else {
               splitter, ok := child.(Splitter)
               if ok {
                    new_split_layout := splitter.Split(to_split)
                    if new_split_layout != nil {
                         return new_split_layout
                    }
               }
          }
     }

     return nil
}

func (layout *VerticalLayout) Remove(query Layout) {
	for i, child := range layout.layouts {
		if child == query {
			layout.layouts = append(layout.layouts[:i], layout.layouts[i+1:]...)
			return
		}
	}
}

func (layout *VerticalLayout) GetParentOf(query Layout) (Layout) {
     for _, child := range layout.layouts {
          if child == query {
               return layout
          } else {
               parent := child.GetParentOf(query)
               if parent != nil {
                    return parent
               }
          }
     }

     return nil
}

func (layout *ViewLayout) View() (view *View) {
	return &layout.view
}

func (layout *ViewLayout) Draw(terminal_dimensions Point) {
     if layout.view.buffer != nil {
          layout.view.buffer.Draw(layout.view.rect, layout.view.scroll, terminal_dimensions)
     }
}

func (layout *ViewLayout) CalculateView(view Rect) {
	layout.view.rect = view
}

func (layout *ViewLayout) Find(query Point) Layout {
	if layout.view.rect.Contains(query) {
		return layout
	}

	return nil
}

func (layout *ViewLayout) Remove(query Layout) {
	// this is dumb?
}

func (layout *ViewLayout) GetParentOf(query Layout) (Layout) {
     // this is dumb?
     return nil
}

func (layout *MainLayout) Split(to_split Layout) (Layout){
     if layout == to_split || layout.Layout == to_split {
          if layout.Layout == to_split {
               if splitter, ok := to_split.(Splitter); ok {
                    return splitter.Split(to_split)
               }
          }

          replacement_layout := &VerticalLayout{}
          replacement_layout.layouts = append(replacement_layout.layouts, layout.Layout)
          replacement_layout.layouts = append(replacement_layout.layouts, &ViewLayout{*to_split.View()})
          layout.Layout = replacement_layout
          return replacement_layout.layouts[0]
     }

     new_split_layout := layout.Layout.(Splitter).Split(to_split)
     if new_split_layout != nil {
          return new_split_layout
     } else {
          // TODO: assuming vertical layout for now
          replacement_layout := &VerticalLayout{}
          replacement_layout.layouts = append(replacement_layout.layouts, &ViewLayout{*to_split.View()})
          replacement_layout.layouts = append(replacement_layout.layouts, &ViewLayout{*to_split.View()})
          layout.Layout = replacement_layout
          return replacement_layout.layouts[0]
     }

     return nil
}

func (layout *MainLayout) GetParentOf(query Layout) (Layout) {
     if layout.Layout == query {
          return layout
     }

     return layout.Layout.GetParentOf(query)
}
