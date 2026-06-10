# Build targets for non-Windows environments.

.PHONY: build zip test vet run clean

build:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap ./cmd/lambda

zip: build
	go run ./scripts/zip.go bootstrap function.zip
	rm -f bootstrap

test:
	go test ./...

vet:
	go vet ./...

run:
	go run ./cmd/server

clean:
	rm -f bootstrap function.zip
