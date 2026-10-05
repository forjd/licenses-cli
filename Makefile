VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test dist clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/licenses ./cmd/licenses

test:
	go test ./...

# Cross-compiled archives for every platform, same as a real release.
dist:
	goreleaser release --snapshot --clean

clean:
	rm -rf bin dist
