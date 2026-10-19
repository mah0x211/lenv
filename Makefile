#
# Makefile
# Created by Masatoshi Fukunaga on 21/02/22
#
BUILD_DIR := $(PWD)/build
TEST_BUILD_DIR ?= $(BUILD_DIR)/test
COVERPROFILE ?= coverage.txt
COVERHTML ?= coverage.html
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

# Share coverage counters between the test executable and real CLI subprocesses.
test:
	mkdir -p "$(abspath $(TEST_BUILD_DIR))"
	go build -race -cover -covermode=atomic -o "$(abspath $(TEST_BUILD_DIR))/lenv" .
	go test -c -race -p 1 -cover -covermode=atomic -o "$(abspath $(TEST_BUILD_DIR))/suite.test" .
	@set -eu; \
	coverdir=$$(mktemp -d "$(abspath $(TEST_BUILD_DIR))/coverage.XXXXXX"); \
	GOCOVERDIR="$$coverdir" LENV_TEST_EXECUTABLE="$(abspath $(TEST_BUILD_DIR))/lenv" \
	"$(abspath $(TEST_BUILD_DIR))/suite.test" -test.count=1 -test.timeout=1m \
		-test.gocoverdir="$$coverdir" -test.coverprofile="$(COVERPROFILE)"
	go tool cover -html "$(COVERPROFILE)" -o "$(COVERHTML)"

coverage: test
	go tool cover -func="$(COVERPROFILE)"

clean:
	go clean
	rm -rf $(PWD)/build

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(PWD)/build/lenv -v .
