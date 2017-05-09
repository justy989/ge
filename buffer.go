package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nsf/termbox-go"
	"io"
	"log"
	"os"
	"strings"
	"unicode"
)

// basic buffer interface. TODO elaborate
type Buffer interface {
	// writer interface implementation
	io.Writer
	// reader interface implementation
	io.Reader
	// return a slice of the lines in the buffer
	Lines() []string
	// insert toInsert at the specified index
	InsertLine(lineIndex int, toInsert string) (err error)
	// set the line at the specified ix to newValue
	SetLine(lineIndex int, newValue string) (err error)
	// delete the line at the specified ix
	DeleteLine(lineIndex int) (err error)
	// clears all lines from the buffer
	Clear() (err error)
	SetCursor(location Point) (err error)
	Cursor() (cursor Point)
	Draw(view Rect, scroll Point, terminal_dimensions Point) (err error)
	//Save() (err error)
}

type Undoer interface {
	Buffer
	Undo() (err error)
	Redo() (err error)
	StartChange()
	Commit() (err error)
}

type ChangeType int

const (
	insertLine = iota
	setLine    = iota
	deleteLine = iota
)

type Change struct {
	t        ChangeType
	old      string
	new      string
	location Point
}

type ChangeGroup struct {
	startCursor Point
	changes     []Change
	endCursor   Point
}

// base implementation of the Buffer interface
type BaseBuffer struct {
	lines  []string
	cursor Point
	// we save by writing out the buffer to io.Writer
	//saver io.Writer

	// next point to read (for implementing the read interface)
	readNext Point
}

type UndoBuffer struct {
	Buffer
	changeIndex int
	changes     []ChangeGroup
	nPending    int
	pending     *ChangeGroup
}

func NewUndoBuffer(buffer Buffer) (b *UndoBuffer) {
	return &UndoBuffer{buffer, -1, nil, 0, nil}
}

// 1. start change (note: record cursor here too)
// 2. make changes (insertline, setline, deleteline, appendline)
// 3. commit change (note: record cursor here too)

func (buffer *UndoBuffer) Undo() (err error) {
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

func (buffer *UndoBuffer) Redo() (err error) {
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

func (buffer *UndoBuffer) StartChange() {
	buffer.nPending++
	if buffer.nPending > 1 {
		return
	}
	// record cursor and add marker to indicate start of undo sequence
	buffer.pending = &ChangeGroup{startCursor: buffer.Cursor()}
}

func (buffer *UndoBuffer) Commit() (err error) {
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

func (buffer *UndoBuffer) InsertLine(lineIndex int, toInsert string) (err error) {
	// TODO: bounds checking
	change := Change{insertLine, "", toInsert, Point{0, lineIndex}}
	buffer.Buffer.InsertLine(lineIndex, toInsert)
	buffer.pending.changes = append(buffer.pending.changes, change)
	return nil
}

func (buffer *UndoBuffer) SetLine(lineIndex int, newValue string) (err error) {
	// TODO: bounds checking
	change := Change{setLine, buffer.Lines()[lineIndex], newValue, Point{0, lineIndex}}
	buffer.Buffer.SetLine(lineIndex, newValue)
	buffer.pending.changes = append(buffer.pending.changes, change)
	return nil
}

func (buffer *UndoBuffer) DeleteLine(lineIndex int) (err error) {
	// TODO: bounds checking
	change := Change{deleteLine, buffer.Lines()[lineIndex], "", Point{0, lineIndex}}
	buffer.Buffer.DeleteLine(lineIndex)
	buffer.pending.changes = append(buffer.pending.changes, change)
	return nil
}

func (buffer *BaseBuffer) String() string {
	var b bytes.Buffer
	for _, line := range buffer.lines {
		b.WriteString(line + "\\n")
	}
	return b.String()
}

func (buffer *BaseBuffer) Write(bytes []byte) (int, error) {
	for _, rawLine := range strings.SplitAfter(string(bytes), "\n") {
		toWrite := strings.TrimRight(rawLine, "\n")
		if numLines := len(buffer.lines); numLines == 0 {
			buffer.lines = append(buffer.lines, toWrite)
		} else {
			buffer.lines[numLines-1] = buffer.lines[numLines-1] + toWrite
		}

		if strings.HasSuffix(rawLine, "\n") {
			buffer.lines = append(buffer.lines, "")
		}
	}
	return len(bytes), nil
}

func (buffer *BaseBuffer) Read(bytes []byte) (int, error) {
	// TODO: Implement reader
	return -1, errors.New("not yet implemented")

	// TODO: read starting at buffer.readNext up to len(bytes)
	var nRead int
	for _ = range bytes {
		nRead++
	}

	return nRead, nil
}

// writer interface implementation
// return a slice of the lines in the buffer
func (buffer *BaseBuffer) Lines() []string {
	return buffer.lines
}

func (buffer *BaseBuffer) validateLineIndex(lineIndex int) (err error) {
	if lineIndex < 0 || lineIndex+1 > len(buffer.lines) {
		return errors.New("invalid line index specified")
	}
	return
}

func (buffer *BaseBuffer) validateLocation(location Point) (err error) {
	if err = buffer.validateLineIndex(location.y); err != nil {
		return err
	} else if location.x >= len(buffer.lines[location.y]) {
		return errors.New("invalid x location specified")
	}
	return
}

// insert toInsert at the specified index
func (buffer *BaseBuffer) InsertLine(lineIndex int, toInsert string) (err error) {
	if lineIndex == len(buffer.lines) {
		buffer.lines = append(buffer.lines, toInsert)
	}

	if err = buffer.validateLineIndex(lineIndex); err != nil {
		return
	}

	// make space for new element
	buffer.lines = append(buffer.lines, "")
	// shift elements right
	copy(buffer.lines[lineIndex+1:], buffer.lines[lineIndex:])
	buffer.lines[lineIndex] = toInsert
	return
}

// set the line at the specified ix to newValue
func (buffer *BaseBuffer) SetLine(lineIndex int, newValue string) (err error) {
	if lineIndex == len(buffer.lines) {
		return buffer.InsertLine(lineIndex, newValue)
	}

	if err = buffer.validateLineIndex(lineIndex); err != nil {
		return
	}

	buffer.lines[lineIndex] = newValue
	return
}

// delete the line at the specified ix
func (buffer *BaseBuffer) DeleteLine(lineIndex int) (err error) {
	if err = buffer.validateLineIndex(lineIndex); err != nil {
		return
	}

	if lineIndex == (len(buffer.lines) - 1) {
		// we are deleting the last line
		buffer.lines = buffer.lines[:lineIndex]
	} else {
		buffer.lines = append(buffer.lines[:lineIndex], buffer.lines[lineIndex+1:]...)
	}
	return
}

// clears all lines from the buffer
func (buffer *BaseBuffer) Clear() (err error) {
	buffer.lines = []string{}
	return
}

func (buffer *BaseBuffer) Draw(view Rect, scroll Point, terminal_dimensions Point) (err error) {
	last_row := scroll.y + view.Height()
	if last_row > len(buffer.lines) {
		last_row = len(buffer.lines)
	}

	max_col := scroll.x + view.Width()
	for y, line := range buffer.lines[scroll.y:last_row] {
		if y >= view.Height() {
			break
		}

		last_col := max_col
		if last_col > len(line) {
			last_col = len(line)
		}

		if scroll.x > last_col {
			continue
		}

		final_y := y + view.top
		if final_y >= terminal_dimensions.y {
			break
		}

		for x, ch := range line[scroll.x:last_col] {
			if x >= view.Width() {
				break
			}

			final_x := x + view.left
			if final_x >= terminal_dimensions.x {
				break
			}

			termbox.SetCell(final_x, final_y, ch, termbox.ColorDefault, termbox.ColorDefault)
		}
	}
	return
}

func (buffer *BaseBuffer) SetCursor(location Point) (err error) {
	if err = buffer.validateLocation(location); err != nil {
		return
	}
	buffer.cursor = location
	return
}

func (buffer *BaseBuffer) Cursor() (cursor Point) {
	return buffer.cursor
}

// buffer interface which adds convenience functions to the basic buffer interface
// TODO: should this be exposed?
type EditableBuffer struct {
	Buffer
}

func NewEditableBuffer(buffer Buffer) (b *EditableBuffer) {
	return &EditableBuffer{buffer}
}

// load text from reader into the buffer
func (buffer *EditableBuffer) Load(reader io.Reader) (err error) {
	_, err = io.Copy(buffer, reader)
	if err != nil {
		file, ok := reader.(*os.File)
		if ok {
			info, err := os.Stat(file.Name())
			if err != nil {
				log.Fatalf("os.Stat() error: %v", err)
			}
			if info.IsDir() {
				names, err := file.Readdirnames(0)
				if err != nil {
					log.Fatalf("Readdirnames() error: %v", err)
				}
				for _, filename := range names {
					err = buffer.InsertLine(len(buffer.Lines()), filename)
					if err != nil {
						log.Fatalf("InsertLine() error: %v", err)
					}
				}
			}
		} else {
			log.Fatalf("io.Copy() error: %v", err)
		}
	}
	return
}

// append string to line
func (buffer *EditableBuffer) Append(lineIndex int, toAppend string) (err error) {
	undoer, ok := buffer.Buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	// TODO: validate input
	buffer.SetLine(lineIndex, buffer.Lines()[lineIndex]+toAppend)
	return
}

// prepend string to line
func (buffer *EditableBuffer) Prepend(lineIndex int, toPrepend string) (err error) {
	undoer, ok := buffer.Buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	// TODO: validate input
	buffer.SetLine(lineIndex, toPrepend+buffer.Lines()[lineIndex])
	return
}

// insert string at point
func (buffer *EditableBuffer) Insert(location Point, toInsert string) (err error) {
	undoer, ok := buffer.Buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	if numLines := len(buffer.Lines()); location.y > numLines {
		return errors.New(fmt.Sprintf("Invalid Point %v", location))
	} else if location.y == numLines {
		return buffer.AppendLine(toInsert)
	}

	line := buffer.Lines()[location.y]
	if numCharacters := len(line); location.x > numCharacters {
		return errors.New(fmt.Sprintf("Invalid Point %v", location))
	} else if location.x == numCharacters {
		return buffer.Append(location.y, toInsert)
	}

	buffer.SetLine(location.y, line[:location.x]+toInsert+line[location.x:])
	return
}

// join line with the line following lineIndex and trim whitespace to a single space
func (buffer *EditableBuffer) Join(lineIndex int) (err error) {
	undoer, ok := buffer.Buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	if numLines := len(buffer.Lines()); (lineIndex + 1) > numLines {
		return errors.New(fmt.Sprintf("Invalid lineIndex %d", lineIndex))
	} else if (lineIndex + 1) == numLines {
		// last line. nothing to join
		return
	}

	trimmedLine := strings.TrimRightFunc(buffer.Lines()[lineIndex], unicode.IsSpace)
	trimmedNextLine := strings.TrimLeftFunc(buffer.Lines()[lineIndex+1], unicode.IsSpace)
	newLine := strings.TrimRightFunc(trimmedLine+" "+trimmedNextLine, unicode.IsSpace)

	buffer.SetLine(lineIndex, newLine)
	buffer.DeleteLine(lineIndex + 1)

	return
}

func (buffer *EditableBuffer) DeleteLine(lineIndex int) error {
	undoer, ok := buffer.Buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	return buffer.Buffer.DeleteLine(lineIndex)
}

func (buffer *EditableBuffer) InsertLine(lineIndex int, toInsert string) error {
	undoer, ok := buffer.Buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	return buffer.Buffer.InsertLine(lineIndex, toInsert)
}

func (buffer *EditableBuffer) SetLine(lineIndex int, newValue string) error {
	undoer, ok := buffer.Buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	return buffer.Buffer.SetLine(lineIndex, newValue)
}

func (buffer *EditableBuffer) AppendLine(toInsert string) error {
	undoer, ok := buffer.Buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	return buffer.Buffer.InsertLine(len(buffer.Lines()), toInsert)
}

// clamp point to point to a character on the buffer including the location
// immediately after the end of lines
func (buffer *EditableBuffer) ClampOn(point Point) (p Point) {
	p.y = Clamp(point.y, 0, len(buffer.Lines())-1)
	p.x = Clamp(point.x, 0, len(buffer.Lines()[p.y]))
	return
}

// clamp point to point to a character on the buffer
func (buffer *EditableBuffer) ClampIn(point Point) (p Point) {
	p.y = Clamp(point.y, 0, len(buffer.Lines())-1)
	p.x = Clamp(point.x, 0, len(buffer.Lines()[p.y])-1)
	return
}

// move cursor along buffer by delta
func (buffer *EditableBuffer) MoveCursor(cursor Point, delta Point) Point {
	final := Point{cursor.x + delta.x, cursor.y + delta.y}
	cursor = buffer.ClampOn(final)
	return cursor
}
