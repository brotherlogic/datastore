package main

import "testing"

func TestMax(t *testing.T) {
	if max(int32(12), 4) != int32(12) {
		t.Errorf("Bad max")
	}
}

func TestMin(t *testing.T) {
	if min(int32(12), 4) != 4 {
		t.Errorf("Bad min")
	}
}

func TestFullExtract(t *testing.T) {
	dir, file := extractFilename("this/is")
	if dir != "this/" || file != "is" {
		t.Errorf("Bad extract %v and %v", dir, file)
	}
}
