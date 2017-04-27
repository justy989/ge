package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

type Point struct {
	x int
	y int
}

func (point *Point) String() string {
	return fmt.Sprintf("%d, %d", point.x, point.y)
}

// basic buffer interface. TODO elaborate
type Buffer interface {
	// writer interface implementation
	io.Writer
	// return a slice of the lines in the buffer
	Lines() []string
	// insert toInsert at the specified index
	InsertLine(lineIndex int, toInsert string) (err error)
	// set the line at the specified ix to newValue
	SetLine(lineIndex int, newValue string) (err error)
	// delete the line at the specified ix
	DeleteLine(lineIndex int) (err error)
	// append toInsert to the end of the buffer
	AppendLine(toInsert string) (err error)
	// clears all lines from the buffer
	Clear() (err error)
}

// base implementation of the Buffer interface
type BaseBuffer struct {
	lines []string
}

func (buffer *BaseBuffer) String() string {
	var b bytes.Buffer
	for _, line := range buffer.lines {
		b.WriteString(line + "\\n")
	}
	return b.String()
}

func (buffer *BaseBuffer) Write(bytes []byte) (int, error) {
	// TODO: only make new lines when we see a \n
	buffer.lines = append(buffer.lines, string(bytes))
	return len(bytes), nil
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

// insert toInsert at the specified index
func (buffer *BaseBuffer) InsertLine(lineIndex int, toInsert string) (err error) {
	if lineIndex == len(buffer.lines) {
		return buffer.AppendLine(toInsert)
	}

	if err = buffer.validateLineIndex(lineIndex); err != nil {
		return
	}

	buffer.lines = append(append(buffer.lines[:lineIndex], toInsert), buffer.lines[lineIndex:]...)
	return
}

// set the line at the specified ix to newValue
func (buffer *BaseBuffer) SetLine(lineIndex int, newValue string) (err error) {
	if lineIndex == len(buffer.lines) {
		return buffer.AppendLine(newValue)
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

// append toInsert to the end of the buffer
func (buffer *BaseBuffer) AppendLine(toInsert string) (err error) {
	buffer.lines = append(buffer.lines, toInsert)
	return
}

// clears all lines from the buffer
func (buffer *BaseBuffer) Clear() (err error) {
	buffer.lines = []string{}
	return
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
	// TODO: error handling
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		buffer.AppendLine(string(scanner.Bytes()))
	}
	return
}

// append string to line
func (buffer *EditableBuffer) Append(lineIndex int, toAppend string) (err error) {
	// TODO: validate input
	buffer.SetLine(lineIndex, buffer.Lines()[lineIndex]+toAppend)
	return
}

// prepend string to line
func (buffer *EditableBuffer) Prepend(lineIndex int, toPrepend string) (err error) {
	// TODO: validate input
	buffer.SetLine(lineIndex, toPrepend+buffer.Lines()[lineIndex])
	return
}

// insert string at point
func (buffer *EditableBuffer) Insert(location Point, toInsert string) (err error) {
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
