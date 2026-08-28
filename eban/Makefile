BIN := eban

.PHONY: build install test vet lint fmt vuln modernize check clean

build:
	go build -o $(BIN) ./cmd/eban

install: build
	install -Dm755 $(BIN) $(HOME)/.local/bin/$(BIN)

test:
	go test ./...

vet:
	go vet ./...

lint:
	go tool golangci-lint run

fmt:
	go tool golangci-lint fmt

vuln:
	go tool govulncheck ./...

modernize:
	go tool modernize ./...

check: vet lint test build

clean:
	rm -f $(BIN)
