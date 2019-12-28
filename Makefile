VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"
ARCH_LIST = darwin linux windows

.PHONY: clean start test release

start:
	go run ./main.go

test:
	go test ./pkg/...

lint:
	test -z $$(gofmt -l .)
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
	mkdir -p $@
	cp README.md $@
	cp LICENSE $@
	GOOS=$* GOARCH=amd64 go build $(LDFLAGS) -o $@/hitpoints ./main.go
