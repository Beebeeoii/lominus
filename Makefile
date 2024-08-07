## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## start: run the application
.PHONY: start
start:
	go run main.go