default:
    @just --list

build:
    go build ./...

test:
    go test -race -count=1 ./...

test-v:
    go test -race -count=1 -v ./...

lint:
    go vet ./...

fmt:
    gofmt -w .

fmt-check:
    @test -z "$$(gofmt -l .)" || (gofmt -d . && exit 1)

check: fmt-check lint test build
