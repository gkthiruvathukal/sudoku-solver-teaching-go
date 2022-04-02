package main

import (
	"testing"
)

func TestGSetNew(t *testing.T) {
	s := new(Set[string])
	s.init()
	if s.size() != 0 {
		t.Errorf("size not zero - actual size is %d", s.size())
	}
}

func TestGSetBasic(t *testing.T) {
	s := new(Set[string])
	s.init()
	s.add("a")
	if s.size() != 1 {
		t.Errorf("size not zero - actual size is %d", s.size())
	}
	if !s.contains("a") {
		t.Errorf("cannot find element  %s", "a")
	}
}

func TestGSetInts(t *testing.T) {

	s := new(Set[int])
	s.init()

	for i := 0; i < 10; i++ {
		s.add(i)
	}
	if s.size() < 10 {
		t.Errorf("incorrect size %d", s.size())
	}
}
