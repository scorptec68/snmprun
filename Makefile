PKG = github.com/scorptec68/snmprun
VERSION=1.0.0

build: build-mac build-linux build-win

build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o snmprun-mac -v $(PKG)

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o snmprun-linux-x64 -v $(PKG)

build-win:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o snmprun-win.exe -v $(PKG)
