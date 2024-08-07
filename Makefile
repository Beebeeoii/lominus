## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## build: build the application
.PHONY: build
build:
	go build -v ./...

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cov: run all tests and display coverage
.PHONY: test/cov
test/cov:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## run: run the application
.PHONY: run
start:
	go run main.go