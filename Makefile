BINARY := apig
SOURCES := $(shell find . -name '*.go' -type f | grep -v _examples)

LDFLAGS := -ldflags="-s -w"

GLIDE_VERSION := 0.11.0

.DEFAULT_GOAL := bin/$(BINARY)

bin/$(BINARY): deps $(SOURCES)
	go generate
	go build $(LDFLAGS) -o bin/$(BINARY)

.PHONY: clean
clean:
	rm -fr bin/*
	rm -fr vendor/*

.PHONY: deps
deps: dep
	./dep ensure

dep:
ifeq ($(shell uname),Darwin)
	curl -fL https://github.com/golang/dep/releases/download/v0.3.2/dep-darwin-amd64 -o dep
	chmod +x dep
else
	curl -fL https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 -o dep
	chmod +x dep
endif

.PHONY: install
install:
	go generate
	go install $(LDFLAGS)

.PHONY: test
test:
	go generate
	go test -cover -v ./apig ./command

.PHONY: generation-test
generation-test: bin/$(BINARY)
	script/generation_test.sh
