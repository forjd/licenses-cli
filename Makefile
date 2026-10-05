VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test dist demo clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/licenses ./cmd/licenses

test:
	go test ./...

# Cross-compiled archives for every platform, same as a real release.
dist:
	goreleaser release --snapshot --clean

# Re-record .github/demo.gif from .github/demo.tape (needs Docker).
demo: build
	docker run --rm -v "$(CURDIR):/vhs" ghcr.io/charmbracelet/vhs .github/demo.tape
	docker run --rm -v "$(CURDIR):/vhs" --entrypoint chown ghcr.io/charmbracelet/vhs $(shell id -u):$(shell id -g) /vhs/.github/demo.gif

clean:
	rm -rf bin dist
