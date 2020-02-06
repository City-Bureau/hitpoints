VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"
ARCH_LIST = darwin linux windows

.PHONY: clean install start test format lint release

start:
	go run ./main.go file

install:
	go get github.com/GeertJohan/go.rice
	go get github.com/GeertJohan/go.rice/rice

test:
	go test ./pkg/...

format:
	test -z $$(gofmt -l .)

lint:
	golint -set_exit_status ./...

clean:
	rm -rf release

release: $(patsubst %, release/hitpoints-%-amd64.tar.gz, $(ARCH_LIST))

release/hitpoints-%-amd64.tar.gz: release/hitpoints-%-amd64
	tar -czvf $@ $<

release/hitpoints-windows-amd64.tar.gz: release/hitpoints-windows-amd64
	mv $</hitpoints $</hitpoints.exe
	tar -czvf $@ $<

.PRECIOUS: release/hitpoints-%-amd64
release/hitpoints-%-amd64:
	rice embed-go -i ./pkg/server
	mkdir -p $@
	cp README.md $@
	cp LICENSE $@
	GOOS=$* GOARCH=amd64 go build $(LDFLAGS) -o $@/hitpoints ./main.go
