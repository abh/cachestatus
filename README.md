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

To also calculate checksums, add `-checksum`.