package utils

import (
	"container/list"
)

type Queue struct {
	list *list.List
}

func NewListQueue() *Queue {
	return &Queue{list: list.New()}
}

func (q *Queue) Init() {
	q.list = list.New()
}

func (q *Queue) Push(item interface{}) {
	q.list.PushBack(item)
}

func (q *Queue) Pop() (interface{}, bool) {
	element := q.list.Front()
	if element == nil {
		return nil, false
	}
	q.list.Remove(element)
	return element.Value, true
}

func (q *Queue) Len() int {
	return q.list.Len()
}

func (q *Queue) Peek() (interface{}, bool) {
	element := q.list.Front()
	if element == nil {
		return nil, false
	}

	return element.Value, true
}

func (q *Queue) IsEmpty() bool {
	return q.list.Len() == 0
}
