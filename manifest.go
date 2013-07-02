package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Manifest struct {
	fh   *os.File
	in   chan FileStatus
	quit chan bool
}

type ManifestEntry struct {
	Path         string
	Size         int64
	Checksum     string `json:",omitempty"`
	LastModified time.Time
}

func CreateManifest(path string) (*Manifest, error) {
	fh, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	m := new(Manifest)
	m.fh = fh
	m.in = make(chan FileStatus, 200)
	go m.readQ()
	return m, nil
}

func (m *Manifest) readQ() {
	for {
		entry := new(ManifestEntry)
		select {
		case fs := <-m.in:

			entry.Path = fs.Path
			entry.Checksum = fs.Checksum
			entry.Size = fs.Size
			entry.LastModified = fs.LastModified

			msg, err := json.Marshal(entry)
			if err != nil {
				log.Printf("Error creating json: %s", err)
			}
			msg = append(msg, "\n"...)
			_, err = m.fh.Write(msg)
			if err != nil {
				log.Printf("Error writing to manifest: %s", err)
			}
		case <-m.quit:
			break
		}
	}
}

func (m *Manifest) Close() {
	m.quit <- true
	m.fh.Close()
}
