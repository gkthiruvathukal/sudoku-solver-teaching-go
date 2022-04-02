package main

import "fmt"

type void struct{}

type Set[T comparable] struct {
	set    map[T]void
	member void
}

func getSet[T comparable]() *Set[T] {
	set := new(Set[T])
	set.init()
	return set
}

func (set *Set[T]) init() {
	set.set = make(map[T]void)
}

func (set *Set[T]) add(entry T) {
	set.set[entry] = set.member
}

func (set *Set[T]) remove(entry T) {
	if set.contains(entry) {
		delete(set.set, entry)
	}
}

func (set *Set[T]) size() int {
	return len(set.set)
}

func (set *Set[T]) contains(entry T) bool {
	_, exists := set.set[entry]
	return exists
}

func (set *Set[T]) display() {
	for k := range set.set {
		fmt.Println(k)
	}
}
