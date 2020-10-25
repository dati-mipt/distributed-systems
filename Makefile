all: fmt test

test:
	go test ./...

fmt:
	go fmt ./...
