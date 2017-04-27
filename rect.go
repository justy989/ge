package main

type Rect struct {
     left int
     top int
     right int
     bottom int
}

func NewRectFromPointAndDimension(point Point, dimensions Point) (Rect){
     return Rect{point.x, point.y, point.x + dimensions.x, point.y + dimensions.y}
}

func (rect* Rect) Width() (int){
     return rect.right - rect.left
}

func (rect* Rect) Height() (int){
     return rect.bottom - rect.top
}

func (rect* Rect) Dimensions() (Point){
     return Point{rect.Width(), rect.Height()}
}

func (rect* Rect) Contains(p Point) (bool){
     if p.x >= rect.left && p.x <= rect.right &&
        p.y >= rect.top && p.y <= rect.bottom {
          return true;
     }

     return false;
}
