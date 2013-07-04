DROPBOX=~/Dropbox/Public/geodns

all:

sh: $(DROPBOX)/ccupdate.sh $(DROPBOX)/ccboot.sh

linux: sh
	GOOS=linux \
	GOARCH=amd64 \
	go build -o $(DROPBOX)/cachestatus
	@echo "curl -sk https://dl.dropboxusercontent.com/u/25895/geodns/ccboot.sh | sh"
	@echo "curl -sk https://dl.dropboxusercontent.com/u/25895/geodns/ccupdate.sh | sh"

$(DROPBOX)/ccupdate.sh: ccupdate.sh
	cp ccupdate.sh $(DROPBOX)

$(DROPBOX)/ccboot.sh: ccboot.sh
	cp ccboot.sh $(DROPBOX)
