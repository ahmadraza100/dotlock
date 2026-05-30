GOCMD := go

build:
	$(GOCMD) build ./cmd/dotlock

test:
	$(GOCMD) test -race -count=1 ./...

coverage:
	$(GOCMD) test -coverprofile=coverage.out ./... && $(GOCMD) tool cover -html=coverage.out

lint:
	golangci-lint run

fmt:
	gofmt -w .

clean:
	rm -f dotlock

install:
	$(GOCMD) install ./cmd/dotlock

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean
