GOHOSTOS := $(shell go env GOHOSTOS)
GOPATH := $(shell go env GOPATH)
VERSION := $(shell git describe --tags --always)

.PHONY: init
# init env
init:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/bufbuild/buf/cmd/buf@latest

.PHONY: config
# genreate internal proto
config:
	buf generate --template buf.gen.config.yaml

	