DROPBOX=~/Dropbox/Public/geodns/cachestatus

all:

sh: $(DROPBOX)/ccupdate.sh

linux: sh
	GOOS=linux \
	GOARCH=amd64 \
	go build -o $(DROPBOX)/cachestatus-linux-x86_64
	@echo "curl -sk https://dl.dropboxusercontent.com/u/25895/geodns/cachestatus/ccboot.sh | sh"
	@echo "curl -sk https://dl.dropboxusercontent.com/u/25895/geodns/cachestatus/ccupdate.sh | sh"

$(DROPBOX)/ccupdate.sh: ccupdate.sh
	cp ccupdate.sh $(DROPBOX)

push:
	( cd $(DROPBOX); sh ../push )
