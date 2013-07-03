# CacheStatus

Tool for testing cache contents

## Input files

The tool supports text files with a file per line or a special json manifest file (see below).

The input file will be fetched over http and the URL is specified with the -filelist parameter.

The text files can be in the format

    sha256  ./path/to/file

or just a file path per line.

## Manifest files

To also test Last-Modified headers and file sizes you can create a special manifest file with
the -createmanifest=/some/path option. You still need the filelist option, so use this to
expand a simple list of files to also have expected sizes, checksums etc.

For example

    ./cachestatus \
    	-createmanifest=/tmp/files.json \
    	-filelist=http://localhost/files.txt \
    	-server origin.example.com

To also calculate checksums, add `-checksum` (recommended).

## Check/prime cache

Example:

	./cachestatus \
	  -filelist http://storage.example.com/files.json \
	  -server localhost \
	  -hostname cdn.example.com \
	  -workers 4 \
	  -checksum

## All options

* -checksum

Check checksums on filse (recommended if testing localhost). A side-effect of doing this is that
on cache misses the program will wait until the file has downloaded before triggering the next
file filling into the cache, too (etc).

Also doing a checksum is the only way to make really really sure the file is right.

* -createmanifest=/path/file.json

Create a json manifest from the (typically text) filelist.

* -filelist=http://...

List of files or manifest to check (see above).

* -hostname=cdn.example.com

Hostname to use in the Host: header (optional).

* -server=origin.example.com

Server to fetch files from (defaults to localhost).

* -workers=n

Set the number of workers (concurrent downloads). The number of CPUs used will be half of what's available in the system or 6, whichever is lower.
