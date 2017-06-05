package edit

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode"
)

// convenience functions for editing using the basic buffer interface

// load text from reader into the buffer
func Load(buffer Buffer, reader io.Reader) (err error) {
	_, err = io.Copy(buffer, reader)
	if err != nil {
		file, ok := reader.(*os.File)
		if ok {
			info, err := os.Stat(file.Name())
			if err != nil {
				log.Fatalf("os.Stat() error: %v", err)
			}
			if info.IsDir() {
				file.Seek(0, os.SEEK_SET)
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
func Append(buffer Buffer, lineIndex int, toAppend string) (err error) {
	undoer, ok := buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	// TODO: validate input
	buffer.SetLine(lineIndex, buffer.Lines()[lineIndex]+toAppend)
	return
}

// prepend string to line
func Prepend(buffer Buffer, lineIndex int, toPrepend string) (err error) {
	undoer, ok := buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	// TODO: validate input
	buffer.SetLine(lineIndex, toPrepend+buffer.Lines()[lineIndex])
	return
}

// insert string at point
func Insert(buffer Buffer, location Point, toInsert string) (err error) {
	undoer, ok := buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	if numLines := len(buffer.Lines()); location.Y > numLines {
		return errors.New(fmt.Sprintf("Invalid Point %v", location))
	} else if location.Y == numLines {
		return AppendLine(buffer, toInsert)
	}

	line := buffer.Lines()[location.Y]
	if numCharacters := len(line); location.X > numCharacters {
		return errors.New(fmt.Sprintf("Invalid Point %v", location))
	} else if location.X == numCharacters {
		return Append(buffer, location.Y, toInsert)
	}

	buffer.SetLine(location.Y, line[:location.X]+toInsert+line[location.X:])
	return
}

// join line with the line following lineIndex and trim whitespace to a single space
func Join(buffer Buffer, lineIndex int) (err error) {
	undoer, ok := buffer.(Undoer)
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

func DeleteLine(buffer Buffer, lineIndex int) error {
	undoer, ok := buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	return buffer.DeleteLine(lineIndex)
}

func InsertLine(buffer Buffer, lineIndex int, toInsert string) error {
	undoer, ok := buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	return buffer.InsertLine(lineIndex, toInsert)
}

func SetLine(buffer Buffer, lineIndex int, newValue string) error {
	undoer, ok := buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	return buffer.SetLine(lineIndex, newValue)
}

func AppendLine(buffer Buffer, toInsert string) error {
	undoer, ok := buffer.(Undoer)
	if ok {
		undoer.StartChange()
		defer undoer.Commit()
	}
	return buffer.InsertLine(len(buffer.Lines()), toInsert)
}

// clamp point to point to a character on the buffer including the location
// immediately after the end of lines
func ClampOn(buffer Buffer, point Point) (p Point) {
	p.Y = Clamp(point.Y, 0, len(buffer.Lines())-1)
	p.X = Clamp(point.X, 0, len(buffer.Lines()[p.Y]))
	return
}

// clamp point to point to a character on the buffer
func ClampIn(buffer Buffer, point Point) (p Point) {
	p.Y = Clamp(point.Y, 0, len(buffer.Lines())-1)
	p.X = Clamp(point.X, 0, len(buffer.Lines()[p.Y])-1)
	return
}

// move cursor along buffer by delta
func MoveCursor(buffer Buffer, cursor Point, delta Point) Point {
	final := Point{cursor.X + delta.X, cursor.Y + delta.Y}
	cursor = ClampOn(buffer, final)
	return cursor
}

// set cursor to buffer location
func SetCursor(buffer Buffer, location Point) error {
	return buffer.SetCursor(location)
}

func Clear(buffer Buffer) error {
	return buffer.Clear()
}
