VERSION := $(shell git describe --tags --always --dirty="-dev")

default: build

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o ./bin/node-healthchecker github.com/flashbots/node-healthchecker/cmd

.PHONY: snapshot
snapshot:
	goreleaser release --snapshot --rm-dist

.PHONY: release
release:
	@rm -rf ./dist
	GITHUB_TOKEN=$$( gh auth token ) goreleaser release
