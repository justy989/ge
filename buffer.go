package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

// basic buffer interface. this interface is intended to provide a minimal set
// of functions which must be implemented to load, display, and manipulate a
// buffer in the go editor; any new methods should be implemented by wrapping
// this interface whenever possible
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
	//MoveCursor(cursor Point, delta Point) (cursor Point) TODO: ADD THIS FOR SURESIES
	//Save() (err error)
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

// generalized function to stringify a buffer
func StringifyBuffer(buffer Buffer) string {
	var b bytes.Buffer
	for _, line := range buffer.Lines() {
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (buffer *BaseBuffer) String() string {
	return StringifyBuffer(buffer)
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
		// allow moving cursor to empty line
		if location.x != 0 {
			return errors.New("invalid x location specified")
		}
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
