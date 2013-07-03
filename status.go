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
	BadSizes     int
	ReadErrors   int

	mu   sync.Mutex
	quit chan bool
}

type WorkerStatus struct {
	Path   string
	Status string
	Mark   byte
}

type FileStatus struct {
	Path         string
	Size         int64
	Checksum     string
	LastModified time.Time
	BadSize      bool
	BadChecksum  bool
	BadRequest   bool
	ReadError    bool
	Miss         bool
}

func NewStatus(nworkers int) *StatusBoard {
	status := new(StatusBoard)
	status.Status = make([]*WorkerStatus, nworkers)
	status.quit = make(chan bool)

	for n := 0; n < nworkers; n++ {
		status.Status[n] = new(WorkerStatus)
	}

	return status
}

func (s *StatusBoard) Quit() {
	s.quit <- true
	close(s.quit)
}

func (s *StatusBoard) Printer() {

	tick := time.Tick(10 * time.Second)

	for {

		statusLine := make([]byte, len(s.Status))

		select {
		case <-s.quit:
			return

		case <-tick:

			s.mu.Lock()

			// terminal.Stdout.Reset()
			// terminal.Stdout.Clear()

			for n, st := range s.Status {
				statusLine[n] = st.Mark
			}

			log.Printf("%s Files: %6d  Misses: %4d  BadRequest: %d  Sizes: %d  Checksums: %d  ReadError: %d\n",
				string(statusLine),
				s.Checks, s.Misses,
				s.BadRequests, s.BadSizes,
				s.BadChecksums,
				s.ReadErrors,
			)

			s.mu.Unlock()

			time.Sleep(4 * time.Second)
		}
	}
}

func (s *StatusBoard) UpdateStatusBoard(id int, path, status string, mark byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(path) > 0 {
		s.Status[id].Path = path
	}

	if len(status) > 0 {
		s.Status[id].Status = status
	}

	s.Status[id].Mark = mark
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
	if fs.BadSize {
		s.BadSizes++
	}
	if fs.ReadError {
		s.ReadErrors++
	}
	if fs.Miss {
		s.Misses++
	}
}
