package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
)

type WorkerGroup struct {
	vhost  *VHost
	in     FileChannel
	status *StatusBoard
	n      int
}

func NewWorkerGroup(vhost *VHost, status *StatusBoard, in FileChannel) *WorkerGroup {
	wg := new(WorkerGroup)
	wg.in = in
	wg.vhost = vhost
	wg.status = status
	return wg
}

func (wg *WorkerGroup) Start() {
	go wg.run(wg.n)
	wg.n++
}

func (wg *WorkerGroup) run(id int) {

	client := &http.Client{}

	for {
		file := <-wg.in
		// log.Printf("%d FILE: %#v\n", id, file)
		if file == nil {
			log.Println(id, "got nil file")
			break
		}

		wg.getFile(id, client, file)
	}
}

func (wg *WorkerGroup) getFile(id int, client *http.Client, file *File) {

	fs := new(FileStatus)
	fs.Path = file.Path
	defer wg.status.AddFileStatus(fs)

	// log.Printf("%d Getting file '%s'\n", id, file.Path)
	wg.status.UpdateStatusBoard(id, file.Path, "GET'ing")

	host := "localhost"
	// host = "flex02.lax04.netdna.com"

	url := "http://" + host + file.Path
	req, err := http.NewRequest("GET", "http://"+host+file.Path, nil)
	if err != nil {
		log.Fatalf("Could not create for %s: %s", url, err)
	}

	req.Host = wg.vhost.Hostname

	// log.Println("REQUEST", req)

	resp, err := client.Do(req)
	if err != nil {
		fs.ReadError = true
		log.Printf("Error fetching %s: %s\n", url, err)
		return
	}
	defer resp.Body.Close()

	wg.status.UpdateStatusBoard(id, file.Path, "Reading response")

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
		if cacheStatus == "" {
			cacheStatus = "[no cache status]"
		}
		log.Printf("%s: %s\n", file.Path, cacheStatus)
	}

	wg.status.UpdateStatusBoard(id, file.Path, "Checking checksum")

	sha := sha256.New()
	size, err := io.Copy(sha, resp.Body)
	if err != nil {
		fs.ReadError = true
		log.Printf("%d Could not read file '%s': %s", id, file.Path, err)
		return
	}

	file.Sha256Actual = hex.EncodeToString(sha.Sum(nil))

	if file.Sha256Actual != file.Sha256Expected {
		fs.BadChecksum = true
		log.Printf("%d Wrong SHA256 for '%s' (size %d)\n", id, file.Path, size)
	} else {
		// log.Println("Ok!")
	}
}
