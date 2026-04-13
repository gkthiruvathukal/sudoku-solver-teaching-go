package main

import (
	"testing"
)

func TestGSetNew(t *testing.T) {
	s := NewSet[string]()
	if s.Len() != 0 {
		t.Errorf("size not zero - actual size is %d", s.Len())
	}
}

func TestGSetBasic(t *testing.T) {
	s := NewSet[string]()
	s.Add("a")
	if s.Len() != 1 {
		t.Errorf("size not zero - actual size is %d", s.Len())
	}
	if !s.Has("a") {
		t.Errorf("cannot find element  %s", "a")
	}
}

func TestGSetInts(t *testing.T) {
	s := NewSet[int]()

	for i := 0; i < 10; i++ {
		s.Add(i)
	}
	if s.Len() < 10 {
		t.Errorf("incorrect size %d", s.Len())
	}
}
