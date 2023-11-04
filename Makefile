.PHONY: test clean build fmt lint prep

GOFILES := $(shell find . -name "*.go")
VERSION := $(shell git describe --tags --abbrev=0)
LDFLAGS := -ldflags '-s -w -X github.com/hrmsk66/terraformify/cmd.version=$(VERSION)'

test:
	go test -v ./...

test-local: test
	cd tests && go test -v -timeout 20m

clean:
	rm -r -f dist/*

build: clean
	go build -trimpath $(LDFLAGS) -o dist/terraformify main.go
	@echo 'To use your locally built version of terraformify, set the PATH with the following command:'
	@echo 'export PATH=$(PWD)/dist:$$PATH'

fmt:
	gofmt -s -w $(GOFILES)

lint:
	golangci-lint run

prep: fmt lint
