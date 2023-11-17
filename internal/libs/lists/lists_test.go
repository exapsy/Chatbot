package lists

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ListsTestSuite struct {
	suite.Suite
}

func (suite *ListsTestSuite) TestNewLinkedList() {
	ll := NewLinkedList(1, 2, 3)
	suite.NotNil(ll)
	suite.Equal(1, ll.value)
	suite.Equal(2, ll.next.value)
	suite.Equal(3, ll.next.next.value)
	suite.Nil(ll.next.next.next)
}

func (suite *ListsTestSuite) TestLinkedListChangeValue() {
	ll := NewLinkedList(1)
	err := ll.ChangeValue(2)
	suite.Nil(err)
	suite.Equal(2, ll.value)
}

func (suite *ListsTestSuite) TestNewDoubleLinkedList() {
	dll := NewDoubleLinkedList(1, 2, 3)
	suite.NotNil(dll)
	suite.Equal(1, dll.value)
	suite.Equal(2, dll.next.value)
	suite.Equal(3, dll.next.next.value)
	suite.Nil(dll.next.next.next)
}

func (suite *ListsTestSuite) TestDoubleLinkedListChangeValue() {
	dll := NewDoubleLinkedList(1)
	err := dll.ChangeValue(2)
	suite.Nil(err)
	suite.Equal(2, dll.value)
}

func (suite *ListsTestSuite) TestNewQueue() {
	q := NewQueue(1, 2, 3)
	suite.NotNil(q)
	suite.Equal(1, q.head.value)
	suite.Equal(3, q.tail.value)
}

func (suite *ListsTestSuite) TestQueueEnqueueDequeue() {
	q := NewQueue[int]()
	suite.NotNil(q)

	// Test Enqueue
	q.Enqueue(1)
	q.Enqueue(2)
	suite.Equal(1, q.head.value)
	suite.Equal(2, q.tail.value)

	// Test Dequeue
	val, err := q.Dequeue()
	suite.Nil(err)
	suite.Equal(1, val)
	suite.Equal(2, q.head.value)
}

func (suite *ListsTestSuite) TestQueueDelete() {
	q := NewQueue[int]()
	suite.NotNil(q)

	// Test deleting from an empty queue
	err := q.Delete(1)
	suite.Error(err, "should error when deleting from an empty queue")

	// Setup queue with elements
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	// Test deleting the head
	err = q.Delete(1)
	suite.NoError(err)
	suite.Equal(2, q.head.value, "head should now be the second element")

	// Test deleting the tail
	q.Enqueue(4) // Add another element
	err = q.Delete(4)
	suite.NoError(err)
	suite.NotEqual(4, q.tail.value, "tail should no longer be the deleted element")

	// Test deleting a middle element
	q.Enqueue(4) // Add it back
	q.Enqueue(5)
	err = q.Delete(4)
	suite.NoError(err)
	suite.Equal(5, q.tail.value, "tail should now be the last element")

	// Test deleting a non-existent element
	err = q.Delete(6)
	suite.Error(err, "should error when deleting a non-existent element")
}

func TestListsTestSuite(t *testing.T) {
	suite.Run(t, new(ListsTestSuite))
}
