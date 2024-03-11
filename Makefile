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
.PHONY: patch
patch:
	./versiontag --force patch

# Build a release
.PHONY: release
release: | clean _release

# Actual snapshot target
.PHONY: _snapshot
_snapshot:
	goreleaser --snapshot --skip-publish --clean --fail-fast

# Actual release target
.PHONY: _release
_release:
	goreleaser release --auto-snapshot --clean

.PHONY: run
run:
	go run main.go

.PHONY: versiontag
versiontag:
	curl -L https://raw.githubusercontent.com/franiglesias/versiontag/master/versiontag \
	-o ./versiontag \
	&& chmod +x ./versiontag \
	&& ./versiontag help
