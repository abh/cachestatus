package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ServerName of the current box
var ServerName string

type VHost struct {
	Hostname         string
	FileListLocation string
	Files            []*File
}

type File struct {
	Path           string
	Sha256Expected string
	Sha256Actual   string
	LastModified   time.Time
	LastChecked    time.Time
	Cached         bool
}

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

type FileChannel chan *File

var Status StatusBoard

type FileStatus struct {
	Path        string
	BadChecksum bool
	BadRequest  bool
	ReadError   bool
	Miss        bool
}

func addStatus(fs *FileStatus) {
	// log.Printf("adding status: %#v\n", fs)
	Status.mu.Lock()
	defer Status.mu.Unlock()
	Status.Checks++

	if fs.BadChecksum {
		Status.BadChecksums++
		Status.BadFiles = append(Status.BadFiles, fs.Path)
	}
	if fs.BadRequest {
		Status.BadRequests++
	}
	if fs.Miss {
		Status.Misses++
	}
}

func statusPrinter() {
	for {

		Status.mu.Lock()

		// terminal.Stdout.Reset()
		// terminal.Stdout.Clear()

		log.Printf("Files: %6d  Misses: %4d  BadRequest: %d  Checksums: %d\n",
			Status.Checks, Status.Misses,
			Status.BadRequests, Status.BadChecksums,
		)

		Status.mu.Unlock()
		time.Sleep(4 * time.Second)
	}
}

func main() {

	var err error

	ServerName, err = os.Hostname()
	if err != nil {
		log.Fatalln("Could not get hostname", err)
	}

	ncpus := runtime.NumCPU()

	ncpus /= 2
	if ncpus > 6 {
		ncpus = 6
	}

	log.Printf("Using up to %d CPUs for sha256'ing\n", ncpus)
	runtime.GOMAXPROCS(ncpus)

	vhost := new(VHost)
	vhost.FileListLocation = "http://storage-hc.dal01.netdna.com/sha256-small.txt"
	vhost.FileListLocation = "http://storage-hc.dal01.netdna.com/sha256.txt"

	vhost.Hostname = "hcinstall.tera-online.com"

	log.Println("Getting file list")
	err = getFileList(vhost)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Got file list")

	workQueue := make(FileChannel)

	nworkers := 10

	Status.Status = make([]*WorkerStatus, nworkers)

	for n := 0; n < nworkers; n++ {
		log.Println("starting worker", n)
		Status.Status[n] = new(WorkerStatus)
		go Worker(n, vhost, workQueue)
	}

	go statusPrinter()

	for i, _ := range vhost.Files {
		// log.Printf("File: %#v\n", file)
		workQueue <- vhost.Files[i]
	}

	for n := 0; n < nworkers; n++ {
		log.Println("closing workers", n)
		workQueue <- nil
	}

	time.Sleep(5 * time.Second)
	log.Println("exiting")

	for _, path := range Status.BadFiles {
		fmt.Println(path)
	}

}

func Worker(id int, vhost *VHost, in FileChannel) {

	client := &http.Client{}

	for {
		file := <-in
		// log.Printf("%d FILE: %#v\n", id, file)
		if file == nil {
			log.Println(id, "got nil file")
			break
		}

		getFile(id, client, vhost.Hostname, file)
	}
}

func getFile(id int, client *http.Client, hostname string, file *File) {

	fs := new(FileStatus)
	fs.Path = file.Path
	defer addStatus(fs)

	// log.Printf("%d Getting file '%s'\n", id, file.Path)
	updateStatus(id, file.Path, "GET'ing")

	host := "localhost"
	// host = "flex02.lax04.netdna.com"

	url := "http://" + host + file.Path
	req, err := http.NewRequest("GET", "http://"+host+file.Path, nil)
	if err != nil {
		log.Fatalf("Could not create for %s: %s", url, err)
	}

	req.Host = hostname

	// log.Println("REQUEST", req)

	resp, err := client.Do(req)
	if err != nil {
		fs.ReadError = true
		log.Printf("Error fetching %s: %s\n", url, err)
		return
	}
	defer resp.Body.Close()

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

func updateStatus(id int, path, status string) {
	Status.mu.Lock()
	defer Status.mu.Unlock()

	if len(path) > 0 {
		Status.Status[id].Current = path
	}

	if len(status) > 0 {
		Status.Status[id].Status = status
	}
}

func getFileList(vhost *VHost) error {
	url := vhost.FileListLocation
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Could not get url %v: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Could not get file list: %d %s", resp.StatusCode, err)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		shaPath := strings.SplitN(scanner.Text(), "  .", 2)
		file := new(File)
		file.Sha256Expected = shaPath[0]
		file.Path = shaPath[1]

		vhost.Files = append(vhost.Files, file)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil

}
