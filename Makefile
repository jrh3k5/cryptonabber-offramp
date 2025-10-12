build:
	go build -o dist/cryptonabber-offramp cmd/main.go

fmt:
	go fmt ./...
	go run golang.org/x/tools/cmd/goimports@latest -w -local github.com/jrh3k5/cryptonabber-offramp .

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