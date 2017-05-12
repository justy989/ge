package main

// the undoer interface wraps a buffer with undo functionality
type Undoer interface {
	Buffer
	Undo() (err error)
	Redo() (err error)
	StartChange()
	Commit() (err error)
}

// internal type which wraps a buffer with undo functionality
type undoBuffer struct {
	Buffer
	changeIndex int
	changes     []ChangeGroup
	nPending    int
	pending     *ChangeGroup
}

// add undo to the provided buffer
func NewUndoer(buffer Buffer) Undoer {
	return &undoBuffer{buffer, -1, nil, 0, nil}
}

// 1. start change (note: record cursor here too)
// 2. make changes (insertline, setline, deleteline, appendline)
// 3. commit change (note: record cursor here too)

func (buffer *undoBuffer) Undo() (err error) {
	if buffer.changeIndex < 0 {
		// nothing to undo
		return nil
	}

	undoGroup := &buffer.changes[buffer.changeIndex]
	for i := len(undoGroup.changes) - 1; i >= 0; i-- {
		toUndo := &undoGroup.changes[i]
		switch toUndo.t {
		default:
			panic("I AM SO FREAKING OUT")
		case insertLine:
			buffer.Buffer.DeleteLine(toUndo.location.y)
		case setLine:
			buffer.Buffer.SetLine(toUndo.location.y, toUndo.old)
		case deleteLine:
			buffer.Buffer.InsertLine(toUndo.location.y, toUndo.old)
		}
	}
	if buffer.changeIndex >= 0 {
		buffer.changeIndex--
	}
	return nil
}

func (buffer *undoBuffer) Redo() (err error) {
	if (buffer.changeIndex + 1) >= len(buffer.changes) {
		// nothing to redo
		return nil
	}

	redoGroup := &buffer.changes[buffer.changeIndex+1]
	for i := len(redoGroup.changes) - 1; i >= 0; i-- {
		toRedo := &redoGroup.changes[i]
		switch toRedo.t {
		default:
			panic("I AM SO FREAKING OUT")
		case insertLine:
			buffer.Buffer.InsertLine(toRedo.location.y, toRedo.new)
		case setLine:
			buffer.Buffer.SetLine(toRedo.location.y, toRedo.new)
		case deleteLine:
			buffer.Buffer.DeleteLine(toRedo.location.y)
		}
	}
	buffer.changeIndex++
	return nil
}

func (buffer *undoBuffer) StartChange() {
	buffer.nPending++
	if buffer.nPending > 1 {
		return
	}
	// record cursor and add marker to indicate start of undo sequence
	buffer.pending = &ChangeGroup{startCursor: buffer.Cursor()}
}

func (buffer *undoBuffer) Commit() (err error) {
	if buffer.pending == nil || buffer.nPending == 0 {
		panic("no change pending!")
	}
	buffer.nPending--
	if buffer.nPending > 0 {
		// we still have more changes to record before we commit
		return nil
	}

	buffer.pending.endCursor = buffer.Cursor()
	if (buffer.changeIndex + 1) >= len(buffer.changes) {
		buffer.changes = append(buffer.changes, *buffer.pending)
		buffer.changeIndex = len(buffer.changes) - 1
	} else {
		buffer.changeIndex++
		buffer.changes[buffer.changeIndex] = *buffer.pending
		buffer.changes = buffer.changes[:buffer.changeIndex+1]
	}

	buffer.pending = nil

	return nil
}

func (buffer *undoBuffer) InsertLine(lineIndex int, toInsert string) (err error) {
	// TODO: bounds checking
	change := Change{insertLine, "", toInsert, Point{0, lineIndex}}
	buffer.Buffer.InsertLine(lineIndex, toInsert)
	buffer.pending.changes = append(buffer.pending.changes, change)
	return nil
}

func (buffer *undoBuffer) SetLine(lineIndex int, newValue string) (err error) {
	// TODO: bounds checking
	change := Change{setLine, buffer.Lines()[lineIndex], newValue, Point{0, lineIndex}}
	buffer.Buffer.SetLine(lineIndex, newValue)
	buffer.pending.changes = append(buffer.pending.changes, change)
	return nil
}

func (buffer *undoBuffer) DeleteLine(lineIndex int) (err error) {
	// TODO: bounds checking
	change := Change{deleteLine, buffer.Lines()[lineIndex], "", Point{0, lineIndex}}
	buffer.Buffer.DeleteLine(lineIndex)
	buffer.pending.changes = append(buffer.pending.changes, change)
	return nil
}
