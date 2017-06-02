package ge

import (
	"fmt"
	"strings"
	"testing"
)

// base buffer test functions
func TestWrite(t *testing.T) {
	buffer := BaseBuffer{}
	buffer.Write([]byte("test\n"))
	t.Log(buffer.String())
	buffer.Write([]byte("blah\n"))
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

	err = buffer.InsertLine(1, "new")
	if len(buffer.Lines()) != 3 {
		t.Logf("\n%s", buffer.String())
		t.Fatalf("buffer has %d lines. expected 3", len(buffer.Lines()))
	}
	if buffer.Lines()[0] != "test" || buffer.Lines()[1] != "new" || buffer.Lines()[2] != "blah" {
		t.Fatal(buffer.String())
	}
}

func TestSetLine(t *testing.T) {
	buffer := BaseBuffer{}
	var err error
	buffer.InsertLine(0, "test")
	buffer.InsertLine(1, "blah")
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
		t.Fatal("expected error due to invalid input")
	}
}

func TestDeleteLine(t *testing.T) {
	buffer := &BaseBuffer{}
	buffer.Write([]byte("one\ntwo\nthree"))
	if len(buffer.Lines()) != 3 {
		t.Fatal("invalid number of lines")
	}

	var err error
	invalidInputs := []int{3, -1}
	for _, invalid := range invalidInputs {
		t.Logf("testing invalid input %d", invalid)
		err = buffer.DeleteLine(invalid)
		if err == nil {
			t.Fatal("expected to fail with invalid input")
		}
		// verify that we still have five lines
		if len(buffer.Lines()) != 3 {
			t.Fatal("invalid number of lines")
		}
	}
	err = buffer.DeleteLine(0)
	if err != nil {
		t.Fatal(err)
	}

	if len(buffer.Lines()) != 2 || buffer.Lines()[0] != "two" {
		t.Fatalf("unexpected buffer state after delete")
	}

	err = buffer.DeleteLine(1)
	if err != nil {
		t.Fatal(err)
	}

	if len(buffer.Lines()) != 1 || buffer.Lines()[0] != "two" {
		t.Fatalf("unexpected buffer state after delete")
	}

	err = buffer.DeleteLine(0)
	if err != nil {
		t.Fatal(err)
	}

	if len(buffer.Lines()) != 0 {
		t.Fatalf("unexpected buffer state after delete")
	}

	err = buffer.DeleteLine(0)
	if err == nil {
		t.Fatal("expected to fail with invalid input. we should have no lines left!")
	}
}

func BenchmarkSetLine(b *testing.B) {
	buffer := &BaseBuffer{}
	Load(buffer, strings.NewReader("line0\nline1\nline2\nline3"))
	for i := 0; i < b.N; i++ {
		buffer.SetLine(2, "test")
	}
}

// testing editableBuffer
func TestLoad(t *testing.T) {
	buffer := &BaseBuffer{}
	err := Load(buffer, strings.NewReader("line0\nline1\nline2\nline3"))
	if err != nil {
		t.Fatal(err)
	}

	if len(buffer.Lines()) != 4 {
		t.Fatal(buffer)
	}

	for ix, line := range buffer.Lines() {
		t.Log(line)
		if expected := fmt.Sprintf("line%d", ix); line != expected {
			t.Fatalf("Invalid line %s, %s", line, expected)
		}
	}
}

func TestInsert(t *testing.T) {
	buffer := &BaseBuffer{}
	err := Load(buffer, strings.NewReader("line0\nline1\nline2\nline3\n"))
	err = Insert(buffer, Point{0, 0}, "new0")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[0] != "new0line0" {
		t.Log(buffer.Lines()[0])
		t.Fatal(buffer)
	}

	err = Insert(buffer, Point{4, 1}, "new1")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[1] != "linenew11" {
		t.Log(buffer.Lines()[1])
		t.Fatal(buffer)
	}

	err = Insert(buffer, Point{5, 2}, "new2")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[2] != "line2new2" {
		t.Log(buffer.Lines()[2])
		t.Fatal(buffer)
	}

	err = Insert(buffer, Point{6, 3}, "invalid0")
	if err == nil {
		t.Log("inserted at invalid location")
		t.Fatal(buffer)
	}

	// line should not have changed
	if buffer.Lines()[3] != "line3" {
		t.Log(buffer.Lines()[3])
		t.Fatal(buffer)
	}

	err = Insert(buffer, Point{0, 4}, "new4")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[4] != "new4" {
		t.Fatal(buffer)
	}

	err = Insert(buffer, Point{0, 6}, "invalid1")
	if err == nil {
		t.Log("inserted at invalid location")
		t.Fatal(buffer)
	}

	err = Insert(buffer, Point{1, 6}, "invalid2")
	if err == nil {
		t.Log("inserted at invalid location")
		t.Fatal(buffer)
	}
}

func TestJoin(t *testing.T) {
	buffer := &BaseBuffer{}
	err := Load(buffer, strings.NewReader("  line0    \n    line1\n  line2"))
	err = Join(buffer, 0)
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[0] != "  line0 line1" {
		t.Log(buffer.Lines()[0])
		t.Fatal(buffer)
	} else if buffer.Lines()[1] != "  line2" {
		t.Log(buffer.Lines()[1])
		t.Fatal(buffer)
	}

	err = Join(buffer, 1)
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Lines()[1] != "  line2" {
		t.Logf("line[1]: '%s'\n", buffer.Lines()[1])
		t.Fatal(buffer)
	}

	err = Join(buffer, 2)
	if err == nil {
		t.Fatal("Invalid line index should fail")
	}
}
