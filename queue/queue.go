package queue

import (
	"task-queue/store"
)

type Queue struct {
	jobs chan *store.Job
}

func NewQueue(size int) *Queue {
	return &Queue{
		jobs: make(chan *store.Job, size),
	}
}

func (q *Queue) Push(job *store.Job) {
	q.jobs <- job
}

func (q *Queue) Pull() *store.Job {
	return <-q.jobs
}

func (q *Queue) Channel() chan *store.Job {
	return q.jobs
}
