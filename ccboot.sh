#!/bin/sh

curl -sk geodns.bitnames.com/cachestatus/ccupdate.sh  | sh

cachestatus \
  -filelist http://storage-hc.dal01.netdna.com/patch.json \
  -server localhost -hostname patch.enmasse.netdna-cdn.com \
  -workers 4 \
  -checksum

# full hcinstall filelist
#   http://storage-hc.dal01.netdna.com/sha256.txt

# test hcinstall filelist
#   http://storage-hc.dal01.netdna.com/sha256-small.txt