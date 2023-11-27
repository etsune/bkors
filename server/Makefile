GOLANGCI_LINT_VERSION := $(shell golangci-lint --version 2>/dev/null)

all: build lint test

build:
	go install go.uber.org/nilaway/cmd/nilaway

test:
	go test -v -race ./...

cover:
	go test -v -race -coverprofile=cover.out -coverpkg=./... -v ./...
	go tool cover -html=cover.out -o cover.html

lint: golangci-lint tidy-lint

golangci-lint:
ifdef GOLANGCI_LINT_VERSION
	@echo "[lint] $(GOLANGCI_LINT_VERSION)"
else
	$(error "golangci-lint not found, please install it from https://golangci-lint.run/usage/install/#local-installation")
endif
	@echo "[lint] golangci-lint run"
	@golangci-lint run

tidy-lint:
	@echo "[lint] go mod tidy"
	@go mod tidy && \
		git diff --exit-code -- go.mod go.sum || \
		(echo "'go mod tidy' changed files" && false)
