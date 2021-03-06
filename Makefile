export CGO_ENABLED=0
export GO111MODULE=on

.PHONY: build

all: build

build: # @HELP build the source code
build:
	go build -v ./...

test: # @HELP run the unit tests and source code validation
test: build license_check linters
	go test github.com/atomix/go-framework/pkg/...

coverage: # @HELP generate unit test coverage data
coverage: build linters license_check
	go test github.com/atomix/go-framework/... -coverprofile=coverage.out -covermode=count -coverpkg=`go list github.com/atomix/go-framework/...`

linters: # @HELP examines Go source code and reports coding problems
	golangci-lint run

license_check: # @HELP examine and ensure license headers exist
	./build/licensing/boilerplate.py -v

protos:
	docker run -it -v `pwd`:/go/src/github.com/atomix/go-framework \
		-w /go/src/github.com/atomix/go-framework \
		--entrypoint build/bin/compile_protos.sh \
		onosproject/protoc-go:stable
