SOURCE_FILES := $(shell find . -type f -name '*.go')

go-stub-package: $(SOURCE_FILES) go.mod go.sum
	go build -o go-stub-package main.go

.PHONY: clean
clean:
	rm -f go-stub-package

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run


