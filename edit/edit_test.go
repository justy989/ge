package edit

import (
	"os"
	"testing"
)

func TestDirectoryLoad(t *testing.T) {
	buffer := &BaseBuffer{}
	// read the current directory
	curDir, err := os.Open(".")
	if err != nil {
		t.Fatalf("Failed to open directory: %v", err)
	}

	// confirm we are testing loading a directory
	info, err := os.Stat(curDir.Name())
	if err != nil {
		t.Fatalf("Failed to stat current directory to confirm it is a directory: %v", err)
	} else if !info.IsDir() {
		// if we get here we need to devise a different test
		t.Fatal(". is not a directory!")
	}

	dirNames, err := curDir.Readdirnames(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirNames) == 0 {
		t.Fatalf("No file names found in %s", curDir.Name())
	}

	// NOTE: Don't return the seek position to the beginning of the directory so we
	// can test that Load works anyway

	Load(buffer, curDir)

	bufferLines := buffer.Lines()
	t.Logf("dirnames '%v'", dirNames)
	t.Logf("buffer '%v'", buffer.String())
	if len(bufferLines) != len(dirNames) {
		t.Fatal("buffer length mismsatch")
	}
	for i, filename := range dirNames {
		if bufferLines[i] != filename {
			t.Fatalf("line %d: unexpected line EXPECTED( %s ), GOT( %s )", i, dirNames, bufferLines[i])
		}
	}
}
