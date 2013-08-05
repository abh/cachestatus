# CacheStatus

Tool for testing cache contents

## Input files

The tool supports text files with a file per line or a special json manifest file (see below).

The input file will be fetched over http and the URL is specified with the -filelist parameter.

You can also specify a local file url: file:///some/path/file.txt or /some/path/file.txt

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


## License

Copyright (C) 2013 Ask Bjørn Hansen

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
