NAME=github-fresh
VERSION=0.3.0
COMMIT=$(shell git rev-parse --short=7 HEAD)
TIMESTAMP:=$(shell date -u '+%Y-%m-%dT%I:%M:%SZ')

LDFLAGS += -X "main.BuildTime=${TIMESTAMP}"
LDFLAGS += -X "main.BuildSHA=${COMMIT}"
LDFLAGS += -X "main.Version=${VERSION}"

.PHONY: all
all: quality test build docker

.PHONY: quality
quality:
	go vet
	go fmt
	docker run -v ${PWD}:/src -w /src -it golangci/golangci-lint golangci-lint run --enable gocritic --enable gosec --enable golint --enable stylecheck --exclude-use-default=false

.PHONY: test
test:
	go test -race -coverprofile=coverage

.PHONY: clean
clean:
	rm -f ${NAME}*

.PHONY: build
build: clean build-darwin build-linux

build-%:
	GOOS=$* GOARCH=386 go build -ldflags '${LDFLAGS}' -o ${NAME}-$*

# todo: hadolint
# todo: sanity check
.PHONY: docker
docker:
	docker build \
	--build-arg NAME="${NAME}" \
	--build-arg VERSION="${VERSION}" \
	--build-arg COMMIT="${COMMIT}" \
	--build-arg BUILD_DATE="${TIMESTAMP}" \
	--tag ${NAME}:${VERSION} .
