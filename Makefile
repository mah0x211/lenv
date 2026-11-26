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

# Combine coverage from package tests and the real CLI subprocesses.
test:
	mkdir -p "$(abspath $(TEST_BUILD_DIR))"
	@set -eu; \
	work=$$(mktemp -d "$(abspath $(TEST_BUILD_DIR))/coverage.XXXXXX"); \
	trap 'rm -rf "$$work"' EXIT; \
	mkdir "$$work/cli"; \
	packages=$$(go list ./internal/...); \
	coverpkg=$$(printf '%s\n' "$$packages" | awk '!/\/internal\/testutil(\/|$$)/ { printf "%s%s", separator, $$0; separator = "," }'); \
	test -n "$$coverpkg"; \
	go build -race -cover -covermode=atomic -coverpkg="$$coverpkg,./cmd/lenv" -o "$$work/lenv" ./cmd/lenv; \
	LENV_TEST_EXECUTABLE="$$work/lenv" LENV_TEST_COVERDIR="$$work/cli" \
		go test -race -p 1 -count=1 -timeout="$(TEST_TIMEOUT)" \
		-covermode=atomic -coverpkg="$$coverpkg" -coverprofile="$$work/tests.txt" ./internal/...; \
	go tool covdata textfmt -i="$$work/cli" -o="$$work/cli.txt"; \
	awk 'NF == 3 && $$3 > 0 { found = 1; exit } END { exit !found }' "$$work/cli.txt" \
		|| { printf '%s\n' 'No coverage counters from the CLI subprocesses' >&2; exit 1; }; \
	awk 'FNR > 1 { statements[$$1] = $$2; hits[$$1] += $$3 } \
		END { for (block in hits) printf "%s %d %.0f\n", block, statements[block], hits[block] }' \
		"$$work/tests.txt" "$$work/cli.txt" > "$$work/blocks.txt"; \
	{ printf '%s\n' 'mode: atomic'; LC_ALL=C sort "$$work/blocks.txt"; } > "$$work/coverage.txt"; \
	go tool cover -html "$$work/coverage.txt" -o "$(abspath $(COVERHTML))"; \
	mv "$$work/coverage.txt" "$(abspath $(COVERPROFILE))"

coverage: test
	go tool cover -func="$(COVERPROFILE)"

clean:
	go clean
	rm -rf $(PWD)/build

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(PWD)/build/lenv -v ./cmd/lenv
