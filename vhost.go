package main

type VHost struct {
	Hostname         string
	FileListLocation string
	Files            []*File
}
