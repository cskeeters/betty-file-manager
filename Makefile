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

.PHONY: fzf

fzf:
	curl -s https://raw.githubusercontent.com/junegunn/fzf/master/src/algo/algo.go -o fzf/algo/algo.go
	curl -s https://raw.githubusercontent.com/junegunn/fzf/master/src/algo/normalize.go -o fzf/algo/normalize.go
	curl -s https://raw.githubusercontent.com/junegunn/fzf/master/src/util/chars.go -o fzf/util/chars.go
	curl -s https://raw.githubusercontent.com/junegunn/fzf/master/src/util/slab.go -o fzf/util/slab.go
	curl -s https://raw.githubusercontent.com/junegunn/fzf/master/src/util/util.go -o fzf/util/util.go
	sed -i '' 's|github.com/junegunn/fzf/src/util|bfm/fzf/util|g' fzf/algo/algo.go
