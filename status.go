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
	Current string
	Status  string
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
}

func (s *StatusBoard) Printer() {

	tick := time.Tick(5 * time.Second)

	for {

		select {
		case <-s.quit:
			log.Println("StatusBoard got quit signal")
			return

		case <-tick:

			s.mu.Lock()

			// terminal.Stdout.Reset()
			// terminal.Stdout.Clear()

			log.Printf("Files: %6d  Misses: %4d  BadRequest: %d  Sizes: %d  Checksums: %d  ReadError: %d\n",
				s.Checks, s.Misses,
				s.BadRequests, s.BadSizes,
				s.BadChecksums,
				s.ReadErrors,
			)

			for n, st := range s.Status {
				if st.Current == "." {
					continue
				}
				log.Println(n, st.Current, st.Status)
			}

			s.mu.Unlock()

			time.Sleep(4 * time.Second)
		}
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
