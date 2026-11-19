#
# Makefile
# Created by Masatoshi Fukunaga on 21/02/22
#
BUILD_DIR := $(PWD)/build
TEST_BUILD_DIR ?= $(BUILD_DIR)/test
COVERPROFILE ?= coverage.txt
COVERHTML ?= coverage.html
TEST_TIMEOUT ?= 5m
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
	$(GOLINT) run $(LINT_OPT) ./...

# Run every package test and use the real CLI for subprocess tests.
test:
	mkdir -p "$(abspath $(TEST_BUILD_DIR))"
	go build -race -cover -covermode=atomic -coverpkg=./internal/... -o "$(abspath $(TEST_BUILD_DIR))/lenv" ./cmd/lenv
	LENV_TEST_EXECUTABLE="$(abspath $(TEST_BUILD_DIR))/lenv" \
		go test -race -p 1 -count=1 -timeout="$(TEST_TIMEOUT)" \
		-covermode=atomic -coverpkg=./internal/... \
		-coverprofile="$(abspath $(COVERPROFILE))" ./internal/...
	go tool cover -html "$(COVERPROFILE)" -o "$(COVERHTML)"

coverage: test
	go tool cover -func="$(COVERPROFILE)"

clean:
	go clean
	rm -rf $(PWD)/build

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(PWD)/build/lenv -v ./cmd/lenv
