BINARY_NAME := ansible-bender2
GOLANG_CI_LINT_VERSION := v2.12.2

build:
	go build -o bin/$(BINARY_NAME) main.go

clean:
	rm -rf bin/*

test: build
	BINARY=$(CURDIR)/bin/$(BINARY_NAME) bats tests/

lint:
	go run "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANG_CI_LINT_VERSION)" run
