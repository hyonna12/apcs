package queue

import "container/list"

type Queue struct {
	l *list.List
}

func NewQueue() *Queue {
	return &Queue{list.New()}
}

func (q *Queue) Push(v any) {
	q.l.PushBack(v)
}

func (q *Queue) Pop() any {
	front := q.l.Front()
	if front == nil {
		return nil
	}

	return q.l.Remove(front)
}

func (q *Queue) IsEmpty() bool {
	return q.l.Len() == 0
}

func (q *Queue) Len() int {
	return q.l.Len()
}
