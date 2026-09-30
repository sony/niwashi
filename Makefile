
export VERSION := $(shell git describe --tags --abbrev=0)
export COMMIT := $(shell git rev-parse --short HEAD)
export DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# derived variables
export

ifndef RELEASE
RELEASE_FLAG:=--snapshot
endif

GOLANGCI_LINT_BIN := ./bin/golangci-lint
GOLANGCI_LINT_VERSION := $(shell cat .golangci-lint-version)

TARGET := nwsctl

all: build

##################################
# Build targets
setup: setup-golangci-lint

setup-golangci-lint:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s $(GOLANGCI_LINT_VERSION)

$(GOLANGCI_LINT_BIN):
	@echo "golangci-lint is not installed. Please run 'make setup' to install it."
	@exit 1

# Build binary for local environment
BUILD_LDFLAGS := -X main.commit=$(COMMIT) -X main.date=$(DATE)

define go-build
	go build \
	-ldflags "$(BUILD_LDFLAGS) $(1)" \
	$(2) \
	-o $@ \
	./cmd/nwsctl
endef

build: ${TARGET}

${TARGET}:
	$(call go-build,-X main.version=$(VERSION)-snapshot,)

build-coverage: ${TARGET}-cov

# Build binary instrumented for coverage measurement
${TARGET}-cov:
	$(call go-build,-X main.version=$(VERSION)-snapshot-coverage,-cover)

build-release:
	goreleaser release --clean --skip=publish ${RELEASE_FLAG}

# Build release binary
release: user-guide build-release

clean: coverage-clean test-clean
	go clean
	- rm nwsctl
	make clean -C docs/user-guide
	- rm -rf dist

##################################
# Document targets

# Create all docs
doc: user-guide

# Create user guide
user-guide:
	make -C docs/user-guide

# Not wired into `check` yet: docs/README.md currently has known broken
# links (tracked separately) that would make this fail today.
docs-lint:
	make -C docs lint

##################################
# License targets

# Collect license texts for all dependencies linked into the binary, into ./LICENSES
licenses:
	GOTOOLCHAIN=$(shell go env GOVERSION) GOROOT=$(shell go env GOROOT) go run github.com/google/go-licenses@v1.6.0 save ./cmd/nwsctl --save_path=./LICENSES --force --ignore github.com/sony/niwashi

##################################
# Test targets

# Tests
test-clean:
	go clean -testcache

lint:
	$(GOLANGCI_LINT_BIN) run ./internal/... ./cmd/...

unit-test:
	go test ./internal/... ./cmd/... --cover

e2e-test:
	make -C test/e2e

check: lint unit-test e2e-test

##################################
# Analytics targets
coverage: coverage-report
	go tool cover -func=.covdata/coverage.out | tail -1 | awk '{print "Total Coverage:",$$3}'

coverage-clean:
	- rm -rf .covdata nwsctl-cov

coverage-unit:
	mkdir -p .covdata/unit
	go test -count=1 -cover ./internal/... ./cmd/... -args -test.gocoverdir=$(CURDIR)/.covdata/unit

coverage-e2e: build-coverage
	mkdir -p .covdata/e2e
	NWSCTL_BINARY=$(CURDIR)/${TARGET}-cov GOCOVERDIR=$(CURDIR)/.covdata/e2e make -C test/e2e

coverage-data: coverage-unit coverage-e2e
	mkdir -p .covdata/merged
	go tool covdata merge -i=.covdata/unit,.covdata/e2e -o=.covdata/merged
	go tool covdata textfmt -i=.covdata/merged -o=.covdata/coverage.out

coverage-report: coverage-data
	go tool cover -html=.covdata/coverage.out -o=.covdata/coverage.html
