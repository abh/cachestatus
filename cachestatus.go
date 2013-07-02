package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// ServerName of the current box
var ServerName string

type File struct {
	Path           string
	Sha256Expected string
	Sha256Actual   string
	LastModified   time.Time
	LastChecked    time.Time
	Cached         bool
}

type FileChannel chan *File

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

	status := NewStatus(nworkers)

	w := NewWorkerGroup(vhost, status, workQueue)

	for n := 0; n < nworkers; n++ {
		log.Println("starting worker", n)
		go w.Start()
	}

	go status.Printer()

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

	for _, path := range status.BadFiles {
		fmt.Println(path)
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
