export PATH := $(PATH):/usr/local/go/bin

.PHONY: build test test-race test-short test-gogpu-ui test-legacy-fyne bench vet fmt lint coverage clean ci qa-a0 help test-hys test-xbar test-mnist test-circuits test-shared test-research install-skills test-skills
# Help target - self-documenting Makefile
help:
	@echo "FeCIM Lattice Tools Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build all packages"
	@echo "  make test           Run all tests"
	@echo "  make test-race      Run tests with race detector"
	@echo "  make test-short     Run only short tests"
	@echo "  make test-gogpu-ui  Run zero-CGO tests for the gogpu/ui shell"
	@echo "  make test-legacy-fyne Run opt-in legacy Fyne package tests"
	@echo "  make test-shared    Run tests for shared packages"
	@echo "  make test-hys       Run tests for Module 1 (Hysteresis)"
	@echo "  make test-xbar      Run tests for Module 2 (Crossbar)"
	@echo "  make test-mnist     Run tests for Module 3 (MNIST)"
	@echo "  make test-circuits  Run tests for Module 4 (Circuits)"
	@echo "  make bench          Run all benchmarks"
	@echo "  make vet            Run go vet"
	@echo "  make fmt            Run gofmt"
	@echo "  make lint           Run golangci-lint (if available)"
	@echo "  make coverage       Generate coverage report"
	@echo "  make clean          Clean build artifacts"
	@echo ""

# Module-specific test targets
test-shared:
	$(GO) test ./shared/...

test-hys:
	$(GO) test ./module1-hysteresis/...

test-xbar:
	$(GO) test ./module2-crossbar/...

test-mnist:
	$(GO) test ./module3-mnist/...

test-circuits:
	$(GO) test ./module4-circuits/...

test-research:
	PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -v
	CGO_ENABLED=0 $(GO) test ./cmd/fecim-lattice-tools -run 'TestResearch|TestDispatchResearch|TestRootUsageListsResearch' -count=1

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

test-gogpu-ui:
	CGO_ENABLED=0 $(GO) test ./shared/viewmodel/... ./internal/gogpuapp/... ./internal/gogpuscreenshot/... ./cmd/fecim-lattice-tools ./cmd/fecim-screenshotter

test-legacy-fyne:
	$(GO) test -tags legacy_fyne ./...

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

arch-check:
	@bash scripts/check-architecture.sh

ci: fmt vet test-short arch-check

# Skills (FeCIM agent skills)
install-skills:
	scripts/install-fecim-skills.sh

test-skills:
	scripts/test-install-fecim-skills.sh
