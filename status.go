package main

import (
	"log"
	"sync"
	"time"
)

type StatusBoard struct {
	Status       []*WorkerStatus
	BadFiles     []string
	Checks       int
	Misses       int
	BadRequests  int
	BadChecksums int

	mu sync.Mutex
}

type WorkerStatus struct {
	Current string
	Status  string
}

type FileStatus struct {
	Path        string
	BadChecksum bool
	BadRequest  bool
	ReadError   bool
	Miss        bool
}

func NewStatus(nworkers int) *StatusBoard {
	status := new(StatusBoard)
	status.Status = make([]*WorkerStatus, nworkers)

	for n := 0; n < nworkers; n++ {
		status.Status[n] = new(WorkerStatus)
	}

	return status
}

func (s *StatusBoard) Printer() {
	for {

		s.mu.Lock()

		// terminal.Stdout.Reset()
		// terminal.Stdout.Clear()

		log.Printf("Files: %6d  Misses: %4d  BadRequest: %d  Checksums: %d\n",
			s.Checks, s.Misses,
			s.BadRequests, s.BadChecksums,
		)

		s.mu.Unlock()

		time.Sleep(4 * time.Second)
	}
}

func (s *StatusBoard) UpdateStatusBoard(id int, path, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(path) > 0 {
		s.Status[id].Current = path
	}

	if len(status) > 0 {
		s.Status[id].Status = status
	}
}

func (s *StatusBoard) AddFileStatus(fs *FileStatus) {
	// log.Printf("adding status: %#v\n", fs)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Checks++

	if fs.BadChecksum {
		s.BadChecksums++
		s.BadFiles = append(s.BadFiles, fs.Path)
	}
	if fs.BadRequest {
		s.BadRequests++
	}
	if fs.Miss {
		s.Misses++
	}
}
