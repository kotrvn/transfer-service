.PHONY: check fmt-check vet build test smoke

API_URL ?= http://localhost:8080
GO_PACKAGES := ./cmd/... ./internal/... ./docs/...

check: fmt-check vet build test

fmt-check:
	@files=$$(gofmt -l cmd internal docs) || exit 1; \
	if [ -n "$$files" ]; then \
		printf 'Run gofmt on:\n%s\n' "$$files" >&2; \
		exit 1; \
	fi

vet:
	go vet $(GO_PACKAGES)

build:
	go build $(GO_PACKAGES)

test:
	go test -race -covermode=atomic -coverprofile=coverage.txt $(GO_PACKAGES)

smoke:
	curl --fail --silent --show-error --max-time 10 "$(API_URL)/health"
	@printf '\n'
	curl --fail --silent --show-error --max-time 10 "$(API_URL)/api/v1/trips"
	@printf '\n'
