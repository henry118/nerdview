BINARY := nerdtui
GO := go
GOFLAGS ?=

.PHONY: build clean vet test test-cover

build:
	$(GO) build $(GOFLAGS) -ldflags="-s -w" -o $(BINARY) .

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

test-cover:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

clean:
	rm -f $(BINARY)
