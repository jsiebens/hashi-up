SHELL := bash
LDFLAGS := "-s -w"
.PHONY: all

.PHONY: build
build:
	go build

.PHONY: dist
dist:
	mkdir -p dist
	GOOS=linux go build -ldflags $(LDFLAGS) -o dist/hashi-up
	GOOS=darwin go build -ldflags $(LDFLAGS) -o dist/hashi-up-darwin
	GOOS=linux GOARCH=arm GOARM=6 go build -ldflags $(LDFLAGS) -o dist/hashi-up-arm
	GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -o dist/hashi-up-arm64
	GOOS=windows go build -ldflags $(LDFLAGS) -o dist/hashi-up.exe

.PHONY: compress
compress:
	for f in dist/hashi-up*; do upx $$f; done

.PHONY: hash
hash:
	for f in dist/hashi-up*; do shasum -a 256 $$f > $$f.sha256; done