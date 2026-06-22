package store

import (
	"log"
	"sync"
	"time"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type Job struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Retries   int       `json:"retries"`
}

type Store struct {
	mu   sync.RWMutex
	jobs map[string]*Job
	aof  *AOF
}

func NewStore(aofPath string) (*Store, error) {
	aof, err := NewAOF(aofPath)
	if err != nil {
		return nil, err
	}

	s := &Store{
		jobs: make(map[string]*Job),
		aof:  aof,
	}

	// Replay AOF on startup
	if err := aof.Replay(s); err != nil {
		log.Printf("AOF replay error: %v", err)
	}

	return s, nil
}

func (s *Store) Save(job *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	if err := s.aof.WriteSave(job); err != nil {
		log.Printf("AOF write error: %v", err)
	}
}

func (s *Store) Get(id string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, exists := s.jobs[id]
	return job, exists
}

func (s *Store) UpdateStatus(id string, status Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if job, exists := s.jobs[id]; exists {
		job.Status = status
		job.UpdatedAt = time.Now()
		if err := s.aof.WriteUpdate(id, status); err != nil {
			log.Printf("AOF write error: %v", err)
		}
	}
}

func (s *Store) GetAll() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		all = append(all, job)
	}
	return all
}

func (s *Store) Close() {
	s.aof.Close()
}
