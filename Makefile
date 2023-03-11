.PHONY: test clean build fmt

GOFILES := $(shell find . -name "*.go")
VERSION := $(shell git describe --tags --abbrev=0)
LDFLAGS := -ldflags '-s -w -X github.com/hrmsk66/terraformify/cmd.version=$(VERSION)'

test:
	cd tests && go test -v -timeout 20m

clean:
	rm -r -f dist/*

build: clean
	go build -trimpath $(LDFLAGS) -o dist/terraformify main.go

fmt:
	gofmt -s -w $(GOFILES)
