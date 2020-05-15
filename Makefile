GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=k8senv

.PHONY: list build build-linux clean

list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs
build:
	$(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME).linux.amd64
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME).darwin.amd64
github-release:
	ifndef TAG
	$(error TAG env variable is not set)
	endif
	ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN env variable is not set)
	endif
	github-release release --user take-the-interview --repo k8senv --tag $(TAG)
	github-release upload --user take-the-interview --repo k8senv --tag $(TAG) \
		--name "k8senv.linux.amd64" \
		--file k8senv.linux.amd64
	github-release upload --user take-the-interview --repo k8senv --tag $(TAG) \
		--name "entrypoint.sh" \
		--file entrypoint.sh
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).linux.amd64
	rm -f $(BINARY_NAME).darwin.amd64

all: build-darwin build-linux
