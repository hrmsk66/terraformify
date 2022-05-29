.PHONY: test clean build prep fmt lint

PACKAGES ?=$(shell $(GO) list ./...)
GOFILES := $(shell find . -name "*.go")
TESTFOLDER := $(shell $(GO) list ./... | grep -E 'cmd$$')
VERSION := $(shell git describe --tags --abbrev=0)
LDFLAGS := -ldflags '-s -w -X github.com/hrmsk66/terraformify/cmd.version=$(VERSION)'

test:
	for d in $(TESTFOLDER); do \
		go test -v $$d -timeout 20m; \
	done

clean:
	rm -r -f bin/*

build: clean
	go build -trimpath $(LDFLAGS) -o dist/terraformify main.go

prep: fmt lint

fmt:
	gofmt -s -w $(GOFILES)

lint:
	golangci-lint run