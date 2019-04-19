NAME=github-fresh
VERSION=0.0.0
COMMIT=$(shell git rev-parse --short=7 HEAD)
TIMESTAMP:=$(shell date -u '+%Y-%m-%dT%I:%M:%S%z')

LDFLAGS += -X "main.BuildTime=${TIMESTAMP}"
LDFLAGS += -X "main.BuildSHA=${COMMIT}"
LDFLAGS += -X "main.Version=${VERSION}"

.PHONY: all
all: quality build docker

.PHONY: build
build: build-darwin build-linux

build-%:
	GOOS=$* GOARCH=386 go build -ldflags '${LDFLAGS}' -o ${NAME}-$*

.PHONY: docker
docker:
	docker build \
	--build-arg NAME="${NAME}" \
	--build-arg VERSION="${VERSION}" \
	--build-arg COMMIT="${COMMIT}" \
	--build-arg TIMESTAMP="${TIMESTAMP}" \
	--tag ${NAME} .

.PHONY: quality
quality:
	go vet
	go fmt
