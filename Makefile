all:
	go build ./cmd/server

protos:
	buf generate

test:
	go test ./...

test-update:
	go test ./... -update


