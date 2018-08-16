VERSION=0.3.0
BUILD_FLAGS := -ldflags="-X main.VERSION=$(VERSION)"

all: clean build

clean:
	@echo "> Remove build dir"
	@rm -rf ./build

build:
	@echo "> Build go application"
	go build ${BUILD_FLAGS} -o sb main.go

version:
	@echo $(VERSION)