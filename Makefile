.PHONY: all
all: quality build

.PHONY: build
build:
	go build -o github-fresh

.PHONY: quality
quality: vet format

.PHONY: format
format:
	go fmt

.PHONY: vet
vet:
	go vet
