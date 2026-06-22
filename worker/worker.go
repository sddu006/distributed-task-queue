package worker

import (
	"fmt"
	"task-queue/queue"
	"task-queue/store"
	"time"
)

type WorkerPool struct {
	numWorkers int
	queue      *queue.Queue
	store      *store.Store
	quit       chan struct{}
}

func NewWorkerPool(numWorkers int, q *queue.Queue, s *store.Store) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		queue:      q,
		store:      s,
		quit:       make(chan struct{}),
	}
}

func (wp *WorkerPool) Start() {
	fmt.Printf("Starting %d workers...\n", wp.numWorkers)
	for i := 0; i < wp.numWorkers; i++ {
		go wp.runWorker(i)
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.quit)
}

func (wp *WorkerPool) runWorker(id int) {
	fmt.Printf("Worker %d is ready\n", id)
	for {
		select {
		case job := <-wp.queue.Channel():
			wp.processJob(id, job)
		case <-wp.quit:
			fmt.Printf("Worker %d shutting down\n", id)
			return
		}
	}
}

func (wp *WorkerPool) processJob(workerID int, job *store.Job) {
	fmt.Printf("Worker %d picked up job %s (type: %s)\n", workerID, job.ID, job.Type)

	wp.store.UpdateStatus(job.ID, store.StatusRunning)

	err := executeJob(job)

	if err != nil {
		job.Retries++
		if job.Retries < 3 {
			fmt.Printf("Job %s failed, retrying (%d/3)\n", job.ID, job.Retries)
			wp.store.UpdateStatus(job.ID, store.StatusPending)
			wp.queue.Push(job)
		} else {
			fmt.Printf("Job %s failed after 3 retries\n", job.ID)
			wp.store.UpdateStatus(job.ID, store.StatusFailed)
		}
		return
	}

	wp.store.UpdateStatus(job.ID, store.StatusDone)
	fmt.Printf("Worker %d completed job %s\n", workerID, job.ID)
}

func executeJob(job *store.Job) error {
	switch job.Type {
	case "send_email":
		fmt.Printf("Sending email with payload: %s\n", job.Payload)
		time.Sleep(2 * time.Second)
	case "resize_image":
		fmt.Printf("Resizing image with payload: %s\n", job.Payload)
		time.Sleep(3 * time.Second)
	case "send_notification":
		fmt.Printf("Sending notification with payload: %s\n", job.Payload)
		time.Sleep(1 * time.Second)
	default:
		fmt.Printf("Processing generic job with payload: %s\n", job.Payload)
		time.Sleep(1 * time.Second)
	}
	return nil
}
