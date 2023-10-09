SOURCE_FILES := $(shell find . -type f -name '*.go')

gostubpkg: $(SOURCE_FILES) go.mod go.sum
	go build -o gostubpkg main.go

.PHONY: clean
clean:
	rm -f gostubpkg

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run


