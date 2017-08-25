.PHONY: all deps test bench clean

all: deps

deps:
	@dep ensure

test:
	@go test -v $(shell go list ./... | grep -v vendor)

bench:
	@go test -bench . -benchmem ./benchmark

clean:
	@rm -rf ./vendor
