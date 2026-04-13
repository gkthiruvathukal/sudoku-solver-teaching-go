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
	set.Clear()
	return set
}

func (set *Set[T]) Clear() {
	set.items = make(map[T]struct{})
}

func (set *Set[T]) Add(entry T) {
	set.items[entry] = struct{}{}
}

func (set *Set[T]) AddAll(other *Set[T]) {
	for entry := range other.items {
		set.Add(entry)
	}
}

func (set *Set[T]) Remove(entry T) {
	if set.Has(entry) {
		delete(set.items, entry)
	}
}

func (set *Set[T]) Len() int {
	return len(set.items)
}

func (set *Set[T]) IsEmpty() bool {
	return set.Len() == 0
}

func (set *Set[T]) Has(entry T) bool {
	_, exists := set.items[entry]
	return exists
}

func (set *Set[T]) HasAll(values ...T) bool {
	for _, value := range values {
		if !set.Has(value) {
			return false
		}
	}
	return true
}

func (set *Set[T]) Values() []T {
	values := make([]T, 0, len(set.items))
	for value := range set.items {
		values = append(values, value)
	}
	return values
}

func (set *Set[T]) Equal(other *Set[T]) bool {
	if set.Len() != other.Len() {
		return false
	}
	for value := range set.items {
		if !other.Has(value) {
			return false
		}
	}
	return true
}

func (set *Set[T]) Clone() *Set[T] {
	clone := NewSet[T]()
	clone.AddAll(set)
	return clone
}

func (set *Set[T]) Union(other *Set[T]) *Set[T] {
	result := set.Clone()
	result.AddAll(other)
	return result
}

func (set *Set[T]) Intersection(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for value := range set.items {
		if other.Has(value) {
			result.Add(value)
		}
	}
	return result
}

func (set *Set[T]) Difference(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for value := range set.items {
		if !other.Has(value) {
			result.Add(value)
		}
	}
	return result
}

func (set *Set[T]) IsSubsetOf(other *Set[T]) bool {
	for value := range set.items {
		if !other.Has(value) {
			return false
		}
	}
	return true
}

func (set *Set[T]) Display() {
	for k := range set.items {
		fmt.Println(k)
	}
}
