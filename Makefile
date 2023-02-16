.PHONY: test godoc

test:
	go install ./...
	go test -timeout 10m -v -race ./...
	go vet ./...

godoc:
	scripts/godoc-install-hint.sh
	@echo "\n\nlaunching godoc server. see docs here: http://localhost:6060/pkg/github.com/wavefronthq/wavefront-sdk-go/senders \n\n"
	godoc -http=:6060
