SHELL := bash

BUILD_DIR ?= $(shell mkdir -p build && cd build && pwd)

GO     ?= go
FIND   ?= find
MKDIR  ?= mkdir
LINTER ?= $(BUILD_DIR)/golangci-lint

.EXPORT_ALL_VARIABLES:
.PHONY: build clean test lint fmt vet icon

clean:
	rm -f $(shell $(FIND) . -type f -name '*.coverprofile')

test: fmt vet
	$(GO) test -v -cover ./...

lint: $(LINTER)
	$(GO) mod verify
	$(LINTER) run -v --no-config --deadline=5m

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

$(LINTER):
	$(shell curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BUILD_DIR) v1.25.0)
