SHELL := bash
Version := $(shell git describe --tags --dirty)
# Version := "dev"
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X github.com/jsiebens/hashi-up/cmd.Version=$(Version) -X github.com/jsiebens/hashi-up/cmd.GitCommit=$(GitCommit)"
.PHONY: all

.PHONY: build
build:
	go build -ldflags $(LDFLAGS)

.PHONY: dist
dist:
	mkdir -p dist
	GOOS=linux go build -ldflags $(LDFLAGS) -o dist/hashi-up
	GOOS=darwin go build -ldflags $(LDFLAGS) -o dist/hashi-up-darwin
	GOOS=linux GOARCH=arm GOARM=6 go build -ldflags $(LDFLAGS) -o dist/hashi-up-armhf
	GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -o dist/hashi-up-arm64
	GOOS=windows go build -ldflags $(LDFLAGS) -o dist/hashi-up.exe

.PHONY: hash
hash:
	for f in dist/hashi-up*; do shasum -a 256 $$f > $$f.sha256; done