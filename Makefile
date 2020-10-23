BINARY_NAME=data_uploader
OS=linux
ARCH=amd64
build:
	GOOS=$(OS) GOARCH=$(ARCH) go build -o ./bin/$(BINARY_NAME)_$(OS)_$(ARCH) ./data_uploader.go
