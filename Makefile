GO := go

COVER_OUT  := coverage.out
COVER_HTML := coverage.html

.PHONY: all
all: build

.PHONY: build
build:
	$(GO) build -o bin/kbit-torrent ./cmd

.PHONY: test
test:
	$(GO) test -v ./...

# Unit tests only: -short skips long-running / network-heavy tests.
.PHONY: test-unit
test-unit:
	$(GO) test -short -v ./...

# Run unit tests with coverage and produce both a summary and an HTML report.
# -coverpkg=./... instruments every package in the module so that packages
# exercised transitively through dependencies also appear in the report.
.PHONY: cover
cover:
	$(GO) test -short -race -coverprofile=$(COVER_OUT) -covermode=atomic -coverpkg=./... ./...
	$(GO) tool cover -func=$(COVER_OUT)
	$(GO) tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo ""
	@echo "HTML report written to $(COVER_HTML)"

# Run the full test suite (unit + integration) with coverage and HTML report.
# Integration tests include the real-network tracker tests; they are slow
# but exercise additional code paths not reached in short mode.
.PHONY: cover-integration
cover-integration:
	$(GO) test -race -coverprofile=$(COVER_OUT) -covermode=atomic -coverpkg=./... ./...
	$(GO) tool cover -func=$(COVER_OUT)
	$(GO) tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo ""
	@echo "HTML report written to $(COVER_HTML)"

.PHONY: install
install: build
	install -Dm755 bin/kbit-torrent $(DESTDIR)/usr/bin/kbit-torrent
	install -Dm644 kbit-torrent.1 $(DESTDIR)/usr/share/man/man1/kbit-torrent.1

.PHONY: clean
clean:
	rm -rf bin $(COVER_OUT) $(COVER_HTML)

.PHONY: run
run: build
	./bin/kbit-torrent
