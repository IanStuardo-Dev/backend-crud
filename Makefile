SHELL := /usr/bin/env bash

.PHONY: migrate-up migrate-down migrate-version proto test test-pretty test-install-pretty

GOCACHE := $(CURDIR)/.gocache

migrate-up:
	@go run ./cmd/migrate up

migrate-down:
	@go run ./cmd/migrate down 1

migrate-version:
	@go run ./cmd/migrate version

proto:
	@./scripts/generate-proto.sh

test:
	@GOCACHE="$(GOCACHE)" go test -v ./...

test-pretty:
	@if command -v gotestsum >/dev/null 2>&1; then \
		GOCACHE="$(GOCACHE)" gotestsum --format testname ./...; \
	else \
		echo "gotestsum is not installed; falling back to 'go test -v ./...'"; \
		GOCACHE="$(GOCACHE)" go test -v ./...; \
	fi

test-install-pretty:
	@GOBIN="$(CURDIR)/bin" GOCACHE="$(GOCACHE)" go install gotest.tools/gotestsum@latest
	@echo "installed gotestsum at ./bin/gotestsum"
	@echo "run with: PATH=\"$(CURDIR)/bin:$$PATH\" make test-pretty"
