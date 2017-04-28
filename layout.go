package main

import "github.com/nsf/termbox-go"

type Layout interface {
     View()(*View)
     Draw(terminal_dimensions Point)
     CalculateView(view Rect)
     Find(query Point) (Layout)
     SplitLayout(layout Layout)
     //SplitVertically() (Layout)
}

type VerticalLayout struct {
     view View
     layouts []Layout
}

type ViewLayout struct {
     view View
}

type MainLayout struct {
     layout Layout
     selected* Layout
}

func (layout *VerticalLayout) View()(view *View) {
     return &layout.view
}

func (layout *VerticalLayout) Draw(terminal_dimensions Point) {
     view_width := layout.view.rect.Width()
     for _, child := range layout.layouts {
          child.Draw(terminal_dimensions)
          for i := 0; i < view_width; i++ {
               termbox.SetCell(layout.view.rect.left + i, child.View().rect.bottom, '-', termbox.ColorDefault, termbox.ColorDefault)
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

func (layout *VerticalLayout) Find(query Point) (Layout) {
     for _, child := range layout.layouts {
          matched_layout := child.Find(query)
          if matched_layout != nil {
               return matched_layout
          }
     }

     return nil
}

func (layout *VerticalLayout) SplitLayout(new_layout Layout) {
     if layout == new_layout {
          layout.layouts = append(layout.layouts, &VerticalLayout{view : layout.view})
     }

     // TODO: split layouts
}

func (layout *VerticalLayout) Remove(query Layout) {
     for i, child := range layout.layouts {
          if child == query {
               layout.layouts = append(layout.layouts[:i], layout.layouts[i+1:]...)
               return
          }
     }
}

func (layout *ViewLayout) View() (view *View) {
     return &layout.view
}

func (layout *ViewLayout) Draw(terminal_dimensions Point) {
     layout.view.buffer.Draw(layout.view.rect, layout.view.scroll, terminal_dimensions)
}

func (layout *ViewLayout) CalculateView(view Rect) {
     layout.view.rect = view
}

func (layout *ViewLayout) Find(query Point) (Layout) {
     if layout.view.rect.Contains(query) {
          return layout
     }

     return nil
}

func (layout *ViewLayout) SplitLayout(new_layout Layout) {
     // this is dumb?
}

func (layout *ViewLayout) Remove(query Layout) {
     // this is dumb?
}
