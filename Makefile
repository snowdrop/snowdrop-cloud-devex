PROJECT     := github.com/snowdrop/k8s-supervisor
GITCOMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_FLAGS := -ldflags="-w -X $(PROJECT)/cmd.GITCOMMIT=$(GITCOMMIT) -X $(PROJECT)/cmd.VERSION=$(VERSION)"
GO          ?= go
GOFMT       ?= $(GO)fmt

# go get -u github.com/shurcooL/vfsgen/cmd/vfsgendev
VFSGENDEV   := $(GOPATH)/bin/vfsgendev
PREFIX      ?= $(shell pwd)

all: clean build

clean:
	@echo "> Remove build dir"
	@rm -rf ./build

build: clean
	@echo "> Build go application"
	go build ${BUILD_FLAGS} -o sb main.go

cross: clean
	gox -osarch="darwin/amd64 linux/amd64" -output="dist/bin/{{.OS}}-{{.Arch}}/sb" $(BUILD_FLAGS)

prepare-release: cross
	./scripts/prepare_release.sh

upload: prepare-release
	./scripts/upload_assets.sh

assets: $(VFSGENDEV)
	@echo ">> writing assets"
	cd $(PREFIX)/pkg/template && go generate
	cd $(PREFIX)/pkg/buildpack && go generate

$(VFSGENDEV):
	cd $(PREFIX)/vendor/github.com/shurcooL/vfsgen/ && go install ./cmd/vfsgendev/...

gofmt:
	@echo ">> checking code style"
	@fmtRes=$$($(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print)); \
	if [ -n "$${fmtRes}" ]; then \
		echo "gofmt checking failed!"; echo "$${fmtRes}"; echo; \
		echo "Please ensure you are using $$($(GO) version) for formatting code."; \
		exit 1; \
	fi

version:
	@echo $(VERSION)