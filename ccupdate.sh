#!/bin/sh
mkdir -p ~/bin/
BIN=~/bin/cachestatus
if [ -f $BIN ]; then
  TIMECOND="-z $BIN"
fi
curl -sk $TIMECOND -o $BIN https://dl.dropboxusercontent.com/u/25895/geodns/cachestatus
chmod a+x $BIN
