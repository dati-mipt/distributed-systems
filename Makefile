all: fmt test

test:
	go test ./consistency/...

fmt:
	go fmt ./...
