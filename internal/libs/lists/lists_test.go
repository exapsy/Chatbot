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

func TestListsTestSuite(t *testing.T) {
	suite.Run(t, new(ListsTestSuite))
}
