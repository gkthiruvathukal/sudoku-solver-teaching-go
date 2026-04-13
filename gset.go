package main

import "fmt"

// Generic Set implementation in Go by GKT
// As the generics framework unfolds, I'll probably move this to native Go generic libraries.
// Nevertheless, this shows it is not too hard to build one's own using the native map

type Set[T comparable] struct {
	items map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	set := new(Set[T])
	set.Reset()
	return set
}

func (set *Set[T]) Reset() {
	set.items = make(map[T]struct{})
}

func (set *Set[T]) Add(entry T) {
	set.items[entry] = struct{}{}
}

func (set *Set[T]) Remove(entry T) {
	if set.Has(entry) {
		delete(set.items, entry)
	}
}

func (set *Set[T]) Len() int {
	return len(set.items)
}

func (set *Set[T]) Has(entry T) bool {
	_, exists := set.items[entry]
	return exists
}

func (set *Set[T]) Display() {
	for k := range set.items {
		fmt.Println(k)
	}
}
