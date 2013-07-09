#!/bin/sh
mkdir -p ~/bin/
DEV=0
BIN=~/bin/cachestatus
if [ -f $BIN ]; then
  TIMECOND="-z $BIN"
fi
sysname=`uname -s | awk '{print tolower($0)}'`
arch=`uname -m`
base=http://geodns.bitnames.com/cachestatus
if [ $DEV -eq 1 ]; then
  base=https://dl.dropboxusercontent.com/u/25895/geodns/cachestatus
fi
url=$base/cachestatus-${sysname}-$arch
echo "Fetching $url"
curl -sk $TIMECOND -o $BIN $url
chmod a+x $BIN
