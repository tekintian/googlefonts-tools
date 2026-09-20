VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS = -ldflags="-s -w -X main.AppVersion=$(VERSION)"

.PHONY: tidy build build-mysql build-postgres build-all run dev clean version

version:
	@echo $(VERSION)

tidy:
	go mod tidy

build:
	go build $(LDFLAGS) -o googlefonts-tools .

build-mysql:
	go build -tags mysql $(LDFLAGS) -o googlefonts-tools .

build-postgres:
	go build -tags postgres $(LDFLAGS) -o googlefonts-tools .

build-all:
	go build -tags "mysql,postgres" $(LDFLAGS) -o googlefonts-tools .

run: build
	./googlefonts-tools -s

dev:
	go run $(LDFLAGS) . -s

clean:
	rm -f googlefonts-tools
	rm -rf storage/db/googlefonts.db