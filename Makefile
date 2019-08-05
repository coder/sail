build:
	go build -o sail .

deps:
	go get ./...

install:
	mv sail /usr/local/bin/sail

all: deps build install