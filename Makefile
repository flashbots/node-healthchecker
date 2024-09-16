VERSION := $(shell git describe --tags --always --dirty="-dev" --match "v*.*.*" || echo "development" )
VERSION := $(VERSION:v%=%)

.PHONY: build
build:
	@CGO_ENABLED=0 go build \
			-ldflags "-X main.version=${VERSION}" \
			-o ./bin/node-healthchecker \
		github.com/flashbots/node-healthchecker/cmd

.PHONY: snapshot
snapshot:
	@goreleaser release --snapshot --clean

.PHONY: help
help:
	@printf "\n=====\n\n"
	@go run github.com/flashbots/node-healthchecker/cmd help
	@printf "\n=====\n\n"
	@go run github.com/flashbots/node-healthchecker/cmd serve --help

.PHONY: serve
serve:
	@go run github.com/flashbots/node-healthchecker/cmd serve
