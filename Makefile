.PHONY: bfm

VERSION = $(shell git describe --always --long --dirty)

bfm:
	go build -ldflags="-X main.version=$(VERSION)"
