GOIMPORTS_VERSION := v0.38.0
GOLANGCI_LINT_VERSION := v1.64.8

deps:
	@echo "Checking goimports version..."
	@if ! command -v goimports >/dev/null 2>&1 || \
		! go version -m $$(which goimports) 2>/dev/null | grep -q "golang.org/x/tools\s*$(GOIMPORTS_VERSION)"; then \
		echo "Installing goimports $(GOIMPORTS_VERSION)..."; \
		go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION); \
	else \
		echo "goimports $(GOIMPORTS_VERSION) already installed"; \
	fi
	@echo "Checking golangci-lint version..."
	@if ! command -v golangci-lint >/dev/null 2>&1 || \
		! golangci-lint version 2>/dev/null | grep -q "$(GOLANGCI_LINT_VERSION)"; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	else \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) already installed"; \
	fi

build:
	go build -o dist/cryptonabber-offramp cmd/main.go

fmt:
	go fmt ./...
	goimports -w -local github.com/jrh3k5/cryptonabber-offramp .

lint:
	golangci-lint run

release: release-clean release-build

release-clean:
	rm -rf dist

release-build:
	env GOOS=darwin GOARCH=amd64 go build -o dist/darwin/amd64/cryptonabber-offramp cmd/main.go 
	tar -C dist/darwin/amd64/ -czvf dist/darwin/amd64/osx-x64.tar.gz cryptonabber-offramp
	env GOOS=windows GOARCH=amd64 go build -o dist/windows/amd64/cryptonabber-offramp.exe cmd/main.go 
	(cd dist/windows/amd64 && zip -r - cryptonabber-offramp.exe) > dist/windows/amd64/win-x64.zip