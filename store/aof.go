package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type AOF struct {
	file   *os.File
	writer *bufio.Writer
}

func NewAOF(path string) (*AOF, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return &AOF{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (a *AOF) WriteSave(job *Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(a.writer, "SAVE %s %s\n", job.ID, string(data))
	if err != nil {
		return err
	}
	return a.writer.Flush()
}

func (a *AOF) WriteUpdate(id string, status Status) error {
	_, err := fmt.Fprintf(a.writer, "UPDATE %s %s\n", id, status)
	if err != nil {
		return err
	}
	return a.writer.Flush()
}

func (a *AOF) Replay(s *Store) error {
	_, err := a.file.Seek(0, 0)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(a.file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 2 {
			continue
		}

		command := parts[0]

		switch command {
		case "SAVE":
			if len(parts) < 3 {
				continue
			}
			var job Job
			if err := json.Unmarshal([]byte(parts[2]), &job); err != nil {
				continue
			}
			s.jobs[job.ID] = &job

		case "UPDATE":
			if len(parts) < 3 {
				continue
			}
			id := parts[1]
			status := Status(parts[2])
			if job, exists := s.jobs[id]; exists {
				job.Status = status
			}
		}
	}
	return scanner.Err()
}

func (a *AOF) Close() {
	a.writer.Flush()
	a.file.Close()
}
