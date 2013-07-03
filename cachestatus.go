package main

import (
	"bufio"
	"flag"
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
	Size           int64
	LastModified   time.Time
	LastChecked    time.Time
	Cached         bool
}

type FileChannel chan *File

var VERSION string = "1.1"

var (
	flagListLocation       = flag.String("filelist", "", "URL for filelist or manifest")
	flagCreateManifestPath = flag.String("createmanifest", "", "path for manifest to be created")
	flagServer             = flag.String("server", "localhost", "Server to check")
	flagHostname           = flag.String("hostname", "", "Host header for checks or source for creating manifest")
	flagChecksum           = flag.Bool("checksum", false, "Check (or create) checksums")
	flagWorkers            = flag.Int("workers", 6, "How many concurrent requests to make")
	flagVersion            = flag.Bool("version", false, "Show version")
	flagVerbose            = flag.Bool("verbose", false, "Verbose output")
)

func init() {
	log.SetPrefix("cachestatus ")
	// log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	flag.Parse()

	if *flagVersion {
		fmt.Println("cachestatus", VERSION)
		os.Exit(0)
	}

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

	if *flagVerbose {
		log.Printf("Using up to %d CPUs for sha256'ing\n", ncpus)
	}
	runtime.GOMAXPROCS(ncpus)

}

func main() {

	if len(*flagListLocation) == 0 {
		log.Fatalln("-filelist url option is required")
	}

	vhost := new(VHost)
	vhost.FileListLocation = *flagListLocation
	vhost.Hostname = *flagHostname

	log.Println("Getting file list")
	err := getFileList(vhost)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Got file list")

	workQueue := make(FileChannel)

	nworkers := *flagWorkers

	status := NewStatus(nworkers)

	w := NewWorkerGroup(vhost, *flagServer, status, workQueue)

	if *flagChecksum {
		w.Options.Checksum = true
	}

	if len(*flagCreateManifestPath) > 0 {
		manifest, err := CreateManifest(*flagCreateManifestPath)
		if err != nil {
			log.Fatalf("Could not open manifest '%s': %s", *flagCreateManifestPath, err)
		}
		defer manifest.Close()
		w.SetOutput(manifest.in)
	}

	for n := 0; n < nworkers; n++ {
		go w.Start()
	}

	go status.Printer()

	for i, _ := range vhost.Files {
		// log.Printf("File: %#v\n", file)
		workQueue <- vhost.Files[i]
	}

	for n := 0; n < nworkers; n++ {
		// this isn't buffered so it makes sure each worker is idle
		// before we continue
		workQueue <- nil
	}

	time.Sleep(1 * time.Second)

	for n, st := range status.Status {
		log.Println(n, st.Path, st.Status, st.Mark)
	}

	for _, path := range status.BadFiles {
		fmt.Println(path)
	}

	log.Println(status.String())

	status.Quit()

	log.Println("exiting")
}

func getFileList(vhost *VHost) error {
	url := vhost.FileListLocation
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Could not get url %v: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Could not get file list '%s': %d", url, resp.StatusCode)
	}

	if strings.HasSuffix(url, ".json") {
		files, err := ReadManifest(resp.Body)
		if err != nil {
			log.Fatalf("Error parsing manifest %s: %s", url, err)
		}
		vhost.Files = files
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		shaPath := strings.SplitN(scanner.Text(), "  .", 2)
		file := new(File)
		if len(shaPath) > 1 {
			file.Sha256Expected = shaPath[0]
			file.Path = shaPath[1]
		} else {
			file.Path = shaPath[0]
		}
		if len(file.Path) == 0 {
			continue
		}

		vhost.Files = append(vhost.Files, file)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil

}
