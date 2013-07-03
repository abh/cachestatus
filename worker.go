package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type WorkerGroup struct {
	vhost   *VHost
	in      FileChannel
	server  string
	status  *StatusBoard
	n       int
	Options struct {
		Checksum bool
	}

	out        chan FileStatus
	sendStatus bool
}

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

func NewWorkerGroup(vhost *VHost, server string, status *StatusBoard, in FileChannel) *WorkerGroup {
	wg := new(WorkerGroup)
	wg.in = in
	wg.vhost = vhost
	wg.server = server
	wg.status = status
	return wg
}

func (wg *WorkerGroup) SetOutput(ch chan FileStatus) {
	wg.out = ch
	wg.sendStatus = true
}

func (wg *WorkerGroup) Start() {
	go wg.run(wg.n)
	wg.n++
}

func (wg *WorkerGroup) run(id int) {

	client := &http.Client{}

	wg.status.UpdateStatusBoard(id, ".", "Starting", 'b')

	for {
		file := <-wg.in
		// log.Printf("%d FILE: %#v\n", id, file)
		if file == nil {
			// log.Println(id, "got nil file")
			wg.status.UpdateStatusBoard(id, ".", "Exited", 'x')
			break
		}

		wg.getFile(id, client, file)
	}

	// log.Printf("Worker %d done", id)
}

func (wg *WorkerGroup) getFile(id int, client *http.Client, file *File) {

	fs := new(FileStatus)
	fs.Path = file.Path
	defer func() {
		wg.status.AddFileStatus(fs)
		if wg.sendStatus {
			wg.out <- *fs
		}
		wg.status.UpdateStatusBoard(id, ".", "Idle", '.')
	}()

	// log.Printf("%d Getting file '%s'\n", id, file.Path)
	wg.status.UpdateStatusBoard(id, file.Path, "Making request", 's')

	url := "http://" + wg.server + file.Path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Could not create for %s: %s", url, err)
	}

	if len(wg.vhost.Hostname) > 0 {
		req.Host = wg.vhost.Hostname
	}

	// log.Println("REQUEST", req)

	resp, err := client.Do(req)
	if err != nil {
		fs.ReadError = true
		log.Printf("Error fetching %s: %s\n", url, err)
		return
	}
	defer resp.Body.Close()

	wg.status.UpdateStatusBoard(id, file.Path, "Reading response", 'r')

	// log.Println("File", url)
	// log.Println("Status", resp.StatusCode, resp.Status)
	// log.Println("Headers", resp.Header)

	if resp.StatusCode != 200 {
		fs.BadRequest = true
		log.Printf("No 200 response for %s: %d\n", file.Path, resp.StatusCode)
		return
	}

	cacheStatus := resp.Header.Get("X-Cache")

	if cacheStatus != "HIT" {
		fs.Miss = true
		// if cacheStatus == "" {
		// cacheStatus = "[no cache status]"
		// }
		log.Printf("%s: %s\n", file.Path, cacheStatus)
	}

	if t, err := time.Parse(TimeFormat, resp.Header.Get("Last-Modified")); err == nil {
		fs.LastModified = t
		if !file.LastModified.IsZero() && file.LastModified != t {
			log.Printf("Last-Modified not matching for '%s' (got '%s', expected '%s')",
				file.Path, t, file.LastModified,
			)
		}
	}

	if !wg.Options.Checksum {
		cl := resp.Header.Get("Content-Length")
		if len(cl) > 0 {
			size, err := strconv.Atoi(cl)
			if err != nil {
				log.Printf("Could not parse Content-Length (%s) from '%s': %s", cl, file.Path, err)
			} else {
				fs.Size = int64(size)
				if file.Size > 0 && fs.Size != file.Size {
					fs.BadSize = true
					log.Printf("'%s' has wrong Content-Length (%d, expected %d)\n", file.Path, fs.Size, file.Size)
				}
			}
		} else {
			log.Printf("No Content-Length header for '%s'\n", file.Path)
		}

		return

	}

	wg.status.UpdateStatusBoard(id, file.Path, "Reading response, making checksum", 'r')

	sha := sha256.New()
	size, err := io.Copy(sha, resp.Body)
	if err != nil {
		fs.ReadError = true
		log.Printf("%d Could not read file '%s': %s", id, file.Path, err)
		return
	}
	fs.Size = size
	if file.Size > 0 && fs.Size != file.Size {
		fs.BadSize = true
		log.Printf("'%s' has wrong size (%d, expected %d)\n", file.Path, fs.Size, file.Size)
	}

	fs.Checksum = hex.EncodeToString(sha.Sum(nil))

	// log.Printf("expected checksum for '%s': %s, got %s", file.Path, file.Sha256Expected, fs.Checksum)

	if len(file.Sha256Expected) > 0 && fs.Checksum != file.Sha256Expected {
		fs.BadChecksum = true
		log.Printf("%d Wrong SHA256 for '%s' (size %d)\n", id, file.Path, size)
	} else {
		// log.Println("Ok!")
	}
}
