release-clean:
	rm -rf dist

release-build:
	env GOOS=darwin GOARCH=amd64 go build -o dist/darwin/amd64/cryptonabber-sync cmd/main.go 
	tar -C dist/darwin/amd64/ -czvf dist/darwin/amd64/osx-x64.tar.gz cryptonabber-sync
	env GOOS=windows GOARCH=amd64 go build -o dist/windows/amd64/cryptonabber-sync.exe cmd/main.go 
	(cd dist/windows/amd64 && zip -r - cryptonabber-sync.exe) > dist/windows/amd64/win-x64.zip