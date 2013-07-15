package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

func ReadManifest(body io.Reader) ([]*File, error) {

	files := make([]*File, 0)

	entry := new(ManifestEntry)

	scanner := bufio.NewScanner(body)
	line := 0
	for scanner.Scan() {
		line++
		err := json.Unmarshal(scanner.Bytes(), entry)
		if err != nil {
			err = fmt.Errorf("%s, line %d: '%s'", err, line, scanner.Bytes())
			return nil, err
		}

		file := new(File)
		file.Path = entry.Path
		file.Size = entry.Size
		file.Sha256Expected = entry.Checksum
		file.LastModified = entry.LastModified

		files = append(files, file)
	}
	return files, nil

}

func CreateManifest(path string) (*Manifest, error) {
	fh, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	m := new(Manifest)
	m.fh = fh
	m.in = make(chan FileStatus, 200)
	m.quit = make(chan bool)
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
