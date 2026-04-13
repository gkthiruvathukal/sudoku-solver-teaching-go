package main

import (
	"slices"
	"testing"
)

func TestGSetNew(t *testing.T) {
	s := NewSet[string]()
	if s.Len() != 0 {
		t.Errorf("size not zero - actual size is %d", s.Len())
	}
	if !s.IsEmpty() {
		t.Error("expected new set to be empty")
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

func TestGSetClear(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	s.Add(2)
	s.Clear()

	if !s.IsEmpty() {
		t.Fatal("expected cleared set to be empty")
	}
	if s.Has(1) || s.Has(2) {
		t.Fatal("expected cleared set to remove all elements")
	}
}

func TestGSetValues(t *testing.T) {
	s := NewSet[int]()
	s.Add(3)
	s.Add(1)
	s.Add(2)

	values := s.Values()
	slices.Sort(values)

	expected := []int{1, 2, 3}
	if !slices.Equal(values, expected) {
		t.Fatalf("Values() = %v, want %v", values, expected)
	}
}

func TestGSetCloneAndEqual(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	s.Add(2)

	clone := s.Clone()
	if !clone.Equal(s) {
		t.Fatal("expected clone to equal original")
	}

	clone.Add(3)
	if clone.Equal(s) {
		t.Fatal("expected clone changes not to affect original equality")
	}
	if s.Has(3) {
		t.Fatal("expected clone to be independent from original")
	}
}

func TestGSetAddAllAndHasAll(t *testing.T) {
	s := NewSet[int]()
	other := NewSet[int]()
	other.Add(1)
	other.Add(2)

	s.Add(0)
	s.AddAll(other)

	if !s.HasAll(0, 1, 2) {
		t.Fatal("expected AddAll to merge all elements")
	}
	if s.HasAll(0, 1, 3) {
		t.Fatal("expected HasAll to report missing elements")
	}
}

func TestGSetSetAlgebra(t *testing.T) {
	left := NewSet[int]()
	right := NewSet[int]()

	for _, value := range []int{1, 2, 3} {
		left.Add(value)
	}
	for _, value := range []int{3, 4} {
		right.Add(value)
	}

	union := left.Union(right)
	if !union.Equal(setFromSlice([]int{1, 2, 3, 4})) {
		t.Fatalf("Union() = %v", union.Values())
	}

	intersection := left.Intersection(right)
	if !intersection.Equal(setFromSlice([]int{3})) {
		t.Fatalf("Intersection() = %v", intersection.Values())
	}

	difference := left.Difference(right)
	if !difference.Equal(setFromSlice([]int{1, 2})) {
		t.Fatalf("Difference() = %v", difference.Values())
	}

	if !intersection.IsSubsetOf(left) {
		t.Fatal("expected intersection to be a subset of left")
	}
	if left.IsSubsetOf(right) {
		t.Fatal("expected left not to be a subset of right")
	}
}

func setFromSlice(values []int) *Set[int] {
	s := NewSet[int]()
	for _, value := range values {
		s.Add(value)
	}
	return s
}
