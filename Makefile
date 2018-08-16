VERSION=0.3.0
BUILD_FLAGS := -ldflags="-X main.VERSION=$(VERSION)"

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

version:
	@echo $(VERSION)