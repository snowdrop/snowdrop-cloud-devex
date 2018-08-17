VERSION=0.3.0
PROJECT := github.com/snowdrop/k8s-supervisor
GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
# PKGS := $(shell go list  ./... | grep -v $(PROJECT)/vendor)
BUILD_FLAGS := -ldflags="-w -X $(PROJECT)/cmd.GITCOMMIT=$(GITCOMMIT) -X $(PROJECT)/cmd.VERSION=$(VERSION)"

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
	./scripts/upload_assets.sh github_api_token=34909fe76fe183d69817d045c6ed030700c394f0

version:
	@echo $(VERSION)