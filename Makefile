GOCACHE ?= /private/tmp/itzd-go-cache

.PHONY: test test-race vet build verify vuln

test:
	GOCACHE=$(GOCACHE) go test ./...

test-race:
	GOCACHE=$(GOCACHE) go test -race ./...

vet:
	GOCACHE=$(GOCACHE) go vet ./...

build:
	GOCACHE=$(GOCACHE) go build ./...

vuln:
	GOCACHE=$(GOCACHE) go run golang.org/x/vuln/cmd/govulncheck@latest ./...

verify: test-race vet build
	git diff --check
