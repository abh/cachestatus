DROPBOX=~/Dropbox/Public/geodns

all:

linux: $(DROPBOX)/ccboot.sh
	GOOS=linux \
	GOARCH=amd64 \
	go build -o $(DROPBOX)/cachestatus
	@echo "curl -sk https://dl.dropboxusercontent.com/u/25895/geodns/ccboot.sh | sh"

$(DROPBOX)/ccboot.sh: ccboot.sh
	cp ccboot.sh $(DROPBOX)
