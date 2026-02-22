GO := go

.PHONY: all
all: build

.PHONY: build
build:
	$(GO) build -o bin/kbit-torrent ./cmd

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: clean
clean:
	rm -rf bin

.PHONY: run
run: build
	./bin/kbit-torrent
