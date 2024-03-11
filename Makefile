include ./.env

.PHONY: build
build:
	go build -o bin/gograph main.go

# Clean the build
.PHONY: clean
clean:
	rm -rf bin/
	rm -rf dist/

# Build a release snapshot
.PHONY: snapshot
snapshot: | clean _snapshot

# Build a release
.PHONY: release
release: | clean _release

# Actual snapshot target
.PHONY: _snapshot
_snapshot:
	goreleaser --snapshot --skip-publish --rm-dist --fail-fast

# Actual release target
.PHONY: _release
_release:
	goreleaser release --rm-dist

.PHONY: run
run:
	go run main.go

