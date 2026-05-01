# Makefile para Crom IA API

BINARY_NAME=cromia-api

.PHONY: all build clean run test

all: build

build:
	go build -o bin/$(BINARY_NAME) api/cmd/server/main.go

run:
	./bin/$(BINARY_NAME)

clean:
	rm -rf bin/ cromia mockserver test_data.db* test_*.log cromia.pid cromia.log

test:
	go test ./...

lint:
	go vet ./...
