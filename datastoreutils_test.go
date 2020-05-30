package main

import (
	"context"
	"testing"
)

func TestBasic(t *testing.T) {
	s := InitTest(true, ".testing")
	err := s.runComputation(context.Background())
	if err != nil {
		t.Errorf("Bad run: %v", err)
	}
}
