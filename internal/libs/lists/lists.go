// Package lists includes a list of types for listing any type of items.
// From LinkedLists to Queues.
package lists

import (
	"fmt"
	"reflect"
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
	if len(values) == 0 {
		return nil // Return nil if no values are provided
	}

	root := &LinkedNode[T]{value: values[0]} // Initialize root with the first value
	head := root

	for i := 1; i < len(values); i++ {
		newNode := &LinkedNode[T]{
			value: values[i],
			next:  nil,
		}

		head.next = newNode // Link the new node
		head = newNode      // Move head to the new node
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
	head *DoubleLinkedNode[T]
	tail *DoubleLinkedNode[T]
}

func NewQueue[T any](values ...T) *Queue[T] {
	queue := &Queue[T]{}
	for _, val := range values {
		queue.Enqueue(val)
	}
	return queue
}

func (q *Queue[T]) Enqueue(val T) {
	newNode := &DoubleLinkedNode[T]{value: val}
	if q.tail != nil {
		q.tail.next = newNode
		newNode.previous = q.tail
	}

	q.tail = newNode

	if q.head == nil {
		q.head = newNode
	}
}

// Delete deletes a value from the queue.
// Expect an O(n) performance, but uses reflect.DeepEqual to compare the values since they're of unknown origin.
// reflect.DeepEqual is fixable if we make sure that T always has an "IsEqual(other T)" function at least. (note for the future self)
func (q *Queue[T]) Delete(val T) error {
	if q.head == nil {
		return fmt.Errorf("queue is empty")
	}

	current := q.head
	for current != nil {
		areEqual := reflect.DeepEqual(current.value, val)
		if areEqual {
			// Case 1: Deleting the head node
			if current == q.head {
				q.head = q.head.next
				if q.head != nil {
					q.head.previous = nil
				} else {
					// The queue becomes empty
					q.tail = nil
				}
			} else {
				// Case 2: Deleting a middle or tail node
				current.previous.next = current.next
				if current.next != nil {
					current.next.previous = current.previous
				} else {
					// Case 2b: Deleting the tail node
					q.tail = current.previous
				}
			}
			return nil
		}
		current = current.next
	}

	return fmt.Errorf("value not found in queue")
}

func (q *Queue[T]) Dequeue() (T, error) {
	if q.head == nil {
		var zeroVal T
		return zeroVal, fmt.Errorf("queue is empty")
	}
	dequeuedValue := q.head.value
	q.head = q.head.next
	if q.head == nil {
		q.tail = nil
	} else {
		q.head.previous = nil
	}

	return dequeuedValue, nil
}
