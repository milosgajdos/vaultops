BUILD=go build
CLEAN=go clean
INSTALL=go install
BUILDPATH=./_build
PACKAGES=$(shell go list ./... |grep -v vendor/)

builddir:
	mkdir -p $(BUILDPATH)

install:
	$(INSTALL) ./$(EXDIR)/...

clean:
	rm -rf $(BUILDPATH)
	$(CLEAN)

dep:
	go get ./...

check:
	for pkg in ${PACKAGES}; do \
		go vet $$pkg || exit ; \
	done

test:
	for pkg in ${PACKAGES}; do \
		go test -coverprofile="../../../$$pkg/coverage.txt" -covermode=atomic $$pkg || exit; \
	done

build: builddir
	go build -ldflags="-s -w" -o "$(BUILDPATH)/vaultops"

all: dep check test build

.PHONY: clean
