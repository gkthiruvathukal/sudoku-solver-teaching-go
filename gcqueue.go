package main

import "log"

// CircularQueue[T] is a circular queue

type CircularQueue[T comparable] struct {
	queue                 []T
	storePos, retrievePos int
	count                 int
	default_zero          T
}

func (cq *CircularQueue[T]) show() {
	log.Printf("storePos = %d, retrievePos = %d, queue = ", cq.storePos, cq.retrievePos)
	log.Println(cq)
}

func (cq *CircularQueue[T]) init(size int) {
	cq.queue = make([]T, size)
	cq.storePos = 0
	cq.retrievePos = 0
	cq.count = 0
}
func (cq *CircularQueue[T]) add(s T) int {
	if cq.isFull() {
		return -1
	} else {
		cq.queue[cq.storePos] = s
		cq.storePos = (cq.storePos + 1) % len(cq.queue)
		cq.count++
		return cq.count
	}
}

func (cq *CircularQueue[T]) remove() (int, T) {
	if cq.isEmpty() {
		return -1, cq.default_zero
	} else {
		item := cq.queue[cq.retrievePos]
		cq.retrievePos = (cq.retrievePos + 1) % len(cq.queue)
		cq.count--
		return cq.count, item
	}
}

func (cq *CircularQueue[T]) isFull() bool {
	return cq.count == len(cq.queue)
}

func (cq *CircularQueue[T]) isEmpty() bool {
	return cq.count == 0
}

func (cq *CircularQueue[T]) size() int {
	return cq.count
}

func (cq *CircularQueue[T]) length() int {
	return len(cq.queue)
}
