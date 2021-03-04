#
# Makefile
# Created by Masatoshi Fukunaga on 21/02/22
#
BUILD_DIR := $(PWD)/build
GOPATH:=$(PWD)/build/deps
GOLINT := `which golangci-lint`
LINT_OPT=--issues-exit-code=0 \
		--sort-results \
		--enable-all \
		--tests=false \
		--disable=lll,noctx,gochecknoinits,gochecknoglobals,gocognit,funlen,gosec,forbidigo,wsl,nestif,gomnd,wrapcheck,nlreturn,exhaustivestruct \
		--exclude=ifElseChain

#
# environment variables
#
.EXPORT_ALL_VARIABLES:
PATH:=$(GOPATH)/bin:$(PATH)
GOPATH:=$(GOPATH)

.PHONY: all lint test coverage clean dist build

all: test build

lint:
	$(GOLINT) run $(LINT_OPT) .

test:
	go test -race -count=1 -p 1 -timeout 1m -coverprofile=coverage.txt -covermode=atomic .
	go tool cover -html coverage.txt -o coverage.html

coverage: test
	go tool cover -func=coverage.txt

clean:
	go clean
	rm -rf $(PWD)/build

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(PWD)/build/lenv -v .
