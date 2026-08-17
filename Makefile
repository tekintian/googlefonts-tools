.PHONY: tidy build run clean

tidy:
	go mod tidy

build:
	go build -o googlefonts-tools .

run: build
	./googlefonts-tools -mode server -port 8000

clean:
	rm -f googlefonts-tools
	rm -rf data/googlefonts.db