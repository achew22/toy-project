all:
	go build ./cmd/server

protos:
	buf generate
	# Run buf generate twice because the test client protos in
	# `internal/server/servertest/client` need to first be generated then
	# compiled.
	buf generate

test:
	go test ./...

test-update:
	go test ./... -update


