package main

import "testing"

func TestFullExtract(t *testing.T) {
	dir, file := extractFilename("this/is")
	if dir != "this" || file != "is" {
		t.Errorf("Bad extract %v and %v", dir, file)
	}
}
