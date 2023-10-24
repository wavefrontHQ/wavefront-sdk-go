.PHONY: all test godoc lint lint-fix

all: test lint

test:
	go test -timeout 1m -v -race ./...
	go vet ./...

godoc:
	@scripts/godoc-install-hint.sh
	@echo "\n\nlaunching godoc server. see docs here: http://localhost:6060/pkg/github.com/wavefronthq/wavefront-sdk-go/senders \n\n"
	godoc -http=:6060

lint:
	@scripts/golangci-lint-install-hint.sh
	golangci-lint run

lint-fix:
	@scripts/golangci-lint-install-hint.sh
	golangci-lint run --fix
