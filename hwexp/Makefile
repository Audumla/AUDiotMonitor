# AUDiot Build & Packaging Makefile

VERSION=0.1.0
BINARY=hwexp

.PHONY: all build clean package

all: build package

build:
	go build -o $(BINARY) ./cmd/hwexp/main.go

package:
	mkdir -p dist
	# Build DEB
	VERSION=$(VERSION) nfpm pkg --target dist/hwexp_$(VERSION)_amd64.deb -f packaging/nfpm.yaml
	# Build RPM
	VERSION=$(VERSION) nfpm pkg --target dist/hwexp_$(VERSION)_amd64.rpm -f packaging/nfpm.yaml
	# Build Tarball
	mkdir -p dist/tar
	cp $(BINARY) configs/*.yaml dist/tar/
	tar -czvf dist/hwexp_$(VERSION)_linux_amd64.tar.gz -C dist/tar .

clean:
	rm -f $(BINARY)
	rm -rf dist/
