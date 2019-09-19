DESCRIPTION="Godock Flowdock Client"
NAME=godock

BUILDVERSION=$(shell cat VERSION)
GO_VERSION=$(shell go version)

# Get the git commit
SHA=$(shell git rev-parse --short HEAD)
BUILD_COUNT=$(shell git rev-list --count HEAD)

BUILD_TAG="${BUILD_COUNT}.${SHA}"

CCOS=windows darwin linux
CCARCH=amd64
CCOUTPUT="package/output/{{.OS}}-{{.Arch}}/"

BIN_FILES=$(NAME)d $(NAME)-agent

build: banner lint
	@echo "Building..."
	@mkdir -p bin/
	@go build \
	-ldflags "-X main.Build=${SHA}" \
	-o bin/${NAME} .
	@echo "[*] Done building $(NAME)..."

banner: deps
	@echo "Project:    $(NAME)"
	@echo "Go Version: ${GO_VERSION}"
	@echo "Go Path:    ${GOPATH}"
	@echo

generate:
	@echo "Running go generate..."
	@go generate ./...

lint:
	@go vet  $$(go list ./... | grep -v /vendor/)
	@for pkg in $$(go list ./... |grep -v /vendor/ |grep -v /services/) ; do \
		golint -min_confidence=1 $$pkg ; \
		done

deps:
	@go get -u golang.org/x/lint/golint

tidy-mods:
	@go mod tidy

vendor-mods:
	@go mod vendor

cc:
	@echo "[-] Cross Compiling $(NAME)"
	@mkdir -p dist/
	@echo "[-] $(NAME) linux"
	@GOOS=linux go build -ldflags "-X main.build=${BUILD_TAG}" -o dist/$(NAME)-linux-amd64
	@GOOS=linux GOARCH=386 go build -ldflags "-X main.build=${BUILD_TAG}" -o dist/$(NAME)-linux-386
	@echo "[-] $(NAME) darwin"
	@GOOS=darwin go build -ldflags "-X main.build=${BUILD_TAG}" -o dist/$(NAME)-darwin-amd64

dist: clean banner lint build cc tar
	@echo "[*] Dist Build Done..."

tar:
	@for bin in dist/*; do \
		tar -cvf $$bin.tar.xz -C dist $$(basename $$bin); \
		rm -rf $$bin; \
	done

test:
	go list ./... | xargs -n1 go test

clean:
	@rm -rf bin/
	@rm -rf .$(NAME)-version
	@rm -rf doc/
	@rm -rf package/
	@rm -rf dist/
	@rm -rf bin/
	@rm -rf tmp/


.PHONY: build
