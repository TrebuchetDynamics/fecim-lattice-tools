.PHONY: build test test-race test-short bench vet fmt lint coverage clean ci qa-a0

qa-a0:
	./scripts/qa_a0.sh

# A0 is a deterministic package-level KPI gate using `go test -json`.
# Output example:
#   LIST_TOTAL=72 JSON_TOTAL=72
#   PKG_SUM pass=72 fail=0 skip=0 total=72
# Hard-fails if LIST_TOTAL != JSON_TOTAL (truncation/partial capture).


GO ?= go
GOFMT ?= gofmt
GOLANGCI_LINT ?= golangci-lint

# Optional knobs
BENCH ?= .
BENCH_COUNT ?= 1
COVERAGE_OUT ?= coverage.out
COVERAGE_HTML ?= coverage.html

build:
	$(GO) build ./...

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

test-short:
	$(GO) test -short ./...

bench:
	$(GO) test ./... -run '^$$' -bench '$(BENCH)' -benchmem -count=$(BENCH_COUNT)

vet:
	$(GO) vet ./...

fmt:
	$(GOFMT) -w .

lint:
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run; \
	else \
		echo "$(GOLANGCI_LINT) not found; skipping lint"; \
	fi

coverage:
	$(GO) test ./... -coverprofile=$(COVERAGE_OUT)
	$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

clean:
	rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)

ci: fmt vet test-short
