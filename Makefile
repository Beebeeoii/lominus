## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## run: run the application
.PHONY: run
start:
	go run main.go