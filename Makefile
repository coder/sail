build:
	go build -o sail .

deps:
	go get ./...

install: deps build
	mv sail /usr/local/bin/sail