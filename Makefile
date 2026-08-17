.PHONY: tidy build build-mysql build-postgres build-all run clean

tidy:
	go mod tidy

build:
	go build -ldflags="-s -w" -o googlefonts-tools .

build-mysql:
	go build -tags mysql -ldflags="-s -w" -o googlefonts-tools .

build-postgres:
	go build -tags postgres -ldflags="-s -w" -o googlefonts-tools .

build-all:
	go build -tags "mysql,postgres" -ldflags="-s -w" -o googlefonts-tools .

run: build
	./googlefonts-tools -s

clean:
	rm -f googlefonts-tools
	rm -rf storage/db/googlefonts.db