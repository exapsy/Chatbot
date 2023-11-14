// Package lists includes a list of types for listing any type of items.
// From LinkedLists to Queues.
package lists

import (
	"fmt"
	"sync"
)

var (
	ErrNodeIsNil = fmt.Errorf("node is nil")
)

type LinkedNode[T any] struct {
	m     sync.RWMutex
	value T
	next  *LinkedNode[T]
}

func NewLinkedList[T any](values ...T) *LinkedNode[T] {
	root := &LinkedNode[T]{}
	valuesSize := len(values)
	if valuesSize > 0 {
		root.value = values[0]
	}

	head := root
	for i := 1; i < valuesSize; i++ {
		// could use recursion here but, eh, inefficient
		// better create it on the spot and keep track of the head
		value := values[i]

		newNode := &LinkedNode[T]{
			value: value,
			next:  nil,
		}

		head.next = newNode
		head = newNode
		head = newNode.next // goes to a nil node which will be assigned on the next iteration above
	}

	return root
}

func (n *LinkedNode[T]) ChangeValue(newValue T) error {
	if n == nil {
		return ErrNodeIsNil
	}

	n.m.Lock()
	defer n.m.Unlock()

	n.value = newValue

	return nil
}

type DoubleLinkedNode[T any] struct {
	m              sync.RWMutex
	value          T
	previous, next *DoubleLinkedNode[T]
}

func NewDoubleLinkedList[T any](values ...T) *DoubleLinkedNode[T] {
	root := &DoubleLinkedNode[T]{}
	valuesSize := len(values)
	if valuesSize > 0 {
		root.value = values[0]
	}

	head := root
	for i := 1; i < valuesSize; i++ {
		// could use recursion here but, eh, inefficient
		// better create it on the spot and keep track of the head
		value := values[i]

		newNode := &DoubleLinkedNode[T]{
			value:    value,
			previous: head, // connect to previous node
			next:     nil,
		}

		head.next = newNode // connect next node to previous

		head = newNode
	}

	return root
}

func (n *DoubleLinkedNode[T]) ChangeValue(newValue T) error {
	if n == nil {
		return ErrNodeIsNil
	}

	n.m.Lock()
	defer n.m.Unlock()

	n.value = newValue

	return nil
}

type Queue[T any] struct {
	*DoubleLinkedNode[T]
}

func NewQueue[T any](values ...T) *Queue[T] {
	root := &Queue[T]{}
	valuesSize := len(values)
	if valuesSize > 0 {
		root.value = values[0]
	}

	root.DoubleLinkedNode = NewDoubleLinkedList(values...)

	return root
}

func (q *Queue[T]) Enqueue(val T) {
	panic("not implemented")
}

func (q *Queue[T]) Dequeue(val T) {
	panic("not implemented")
}
