include ./.env

build:
	go build -o bin/gograph main.go

clean:
	rm -rf bin/

snapshot:
	goreleaser --snapshot --skip-publish --rm-dist --fail-fast


.PHONY: run
run:
	go run main.go

