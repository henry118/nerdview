BINARY := nerdtui
GO := go
GOFLAGS ?=

.PHONY: build clean vet test

build:
	$(GO) build $(GOFLAGS) -ldflags="-s -w" -o $(BINARY) .

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

clean:
	rm -f $(BINARY)
