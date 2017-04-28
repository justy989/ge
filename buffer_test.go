package main

import (
	"fmt"
	"strings"
	"testing"
)

// base buffer test functions
func TestWrite(t *testing.T) {
	buffer := BaseBuffer{}
	buffer.Write([]byte("test"))
	t.Log(buffer.String())
	buffer.Write([]byte("blah"))
	if buffer.Lines()[0] != "test" || buffer.Lines()[1] != "blah" {
		t.Fatal(buffer.String())
	}
}

func TestInsertLine(t *testing.T) {
	buffer := BaseBuffer{}
	var err error
	err = buffer.InsertLine(0, "test")
	if err != nil {
		t.Fatal(err)
	}

	err = buffer.InsertLine(1, "blah")
	if buffer.Lines()[0] != "test" || buffer.Lines()[1] != "blah" {
		t.Fatal(buffer.String())
	}
}

func TestSetLine(t *testing.T) {
	buffer := BaseBuffer{}
	var err error
	buffer.AppendLine("test")
	buffer.AppendLine("blah")
	err = buffer.SetLine(0, "new1")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[0] != "new1" {
		t.Fatal(buffer.String())
	}

	err = buffer.SetLine(1, "new2")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[1] != "new2" {
		t.Fatal(buffer.String())
	}

	err = buffer.SetLine(2, "new3")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[2] != "new3" {
		t.Fatal(buffer.String())
	}

	err = buffer.SetLine(6, "new6")
	if err == nil {
		t.FailNow()
	}
}

// testing editableBuffer
func TestLoad(t *testing.T) {
	buffer := NewEditableBuffer(&BaseBuffer{})
	err := buffer.Load(strings.NewReader("line0\nline1\nline2\nline3\n"))
	if err != nil {
		t.Fatal(err)
	}

	if len(buffer.Lines()) != 4 {
		t.Fatal(buffer)
	}

	for ix, line := range buffer.Lines() {
		t.Log(line)
		if expected := fmt.Sprintf("line%d", ix); line != expected {
			t.Fatalf("Invalid line %d, %s", line)
		}
	}
}

func TestInsert(t *testing.T) {
	buffer := NewEditableBuffer(&BaseBuffer{})
	err := buffer.Load(strings.NewReader("line0\nline1\nline2\nline3\n"))
	err = buffer.Insert(Point{0, 0}, "new0")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[0] != "new0line0" {
		t.Log(buffer.Lines()[0])
		t.Fatal(buffer)
	}

	err = buffer.Insert(Point{4, 1}, "new1")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[1] != "linenew11" {
		t.Log(buffer.Lines()[1])
		t.Fatal(buffer)
	}

	err = buffer.Insert(Point{5, 2}, "new2")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[2] != "line2new2" {
		t.Log(buffer.Lines()[2])
		t.Fatal(buffer)
	}

	err = buffer.Insert(Point{6, 3}, "invalid0")
	if err == nil {
		t.Log("inserted at invalid location")
		t.Fatal(buffer)
	}

	// line should not have changed
	if buffer.Lines()[3] != "line3" {
		t.Log(buffer.Lines()[3])
		t.Fatal(buffer)
	}

	err = buffer.Insert(Point{0, 4}, "new4")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[4] != "new4" {
		t.Fatal(buffer)
	}

	err = buffer.Insert(Point{0, 6}, "invalid1")
	if err == nil {
		t.Log("inserted at invalid location")
		t.Fatal(buffer)
	}

	err = buffer.Insert(Point{1, 6}, "invalid2")
	if err == nil {
		t.Log("inserted at invalid location")
		t.Fatal(buffer)
	}
}
