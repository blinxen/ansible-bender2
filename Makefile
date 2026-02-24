BINARY_NAME := ansible-bender2

build:
	go build -o bin/$(BINARY_NAME) main.go

clean:
	rm -rf bin/*

test: build
	BINARY=$(CURDIR)/bin/$(BINARY_NAME) bats tests/
