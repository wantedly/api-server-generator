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
deps: glide
	go get github.com/jteeuwen/go-bindata/...
	./glide install
	
	go get github.com/gin-gonic/gin
	go install github.com/gin-gonic/gin

	go get github.com/jinzhu/gorm
	go install github.com/jinzhu/gorm

	go get github.com/mattn/go-sqlite3
	go install github.com/mattn/go-sqlite3

glide:
ifeq ($(shell uname),Darwin)
	curl -fL https://github.com/Masterminds/glide/releases/download/v$(GLIDE_VERSION)/glide-v$(GLIDE_VERSION)-darwin-amd64.zip -o glide.zip
	unzip glide.zip
	mv ./darwin-amd64/glide ./glide
	rm -fr ./darwin-amd64
	rm ./glide.zip
else
	curl -fL https://github.com/Masterminds/glide/releases/download/v$(GLIDE_VERSION)/glide-v$(GLIDE_VERSION)-linux-amd64.zip -o glide.zip
	unzip glide.zip
	mv ./linux-amd64/glide ./glide
	rm -fr ./linux-amd64
	rm ./glide.zip
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
