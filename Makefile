.PHONY: bfm

VERSION = $(shell git describe --always --long --dirty)

ifeq ($(shell uname -s),Darwin)
	SIGN = codesign --sign - --force --preserve-metadata=entitlements,requirements,flags,runtime /usr/local/bin/bfm
endif

bfm:
	go build -ldflags="-X main.version=$(VERSION)"

install:
	cp bfm /usr/local/bin/
	$(SIGN)
