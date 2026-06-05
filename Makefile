.DEFAULT_GOAL := build

.PHONY: build clean test

build:
	go install ./...

clean:
	go clean ./...
	rm -f ${GOPATH}/bin/*

test:
	go test ./... -count=1
	go vet ./...
