.PHONY: test test-ci test-race test-race-ci bench bench-ci fmt vet

# Default timeouts are intentionally conservative to avoid CI flakes.
GO_TEST_TIMEOUT ?= 10m
GO_TEST_RACE_TIMEOUT ?= 20m

# Base flags used in CI to improve reproducibility.
GO_TEST_FLAGS := -count=1 -shuffle=off -trimpath -timeout=$(GO_TEST_TIMEOUT)

# --- Testing ---

test:
	go test ./...

test-ci:
	GO_TEST_TIMEOUT=$(GO_TEST_TIMEOUT) ./scripts/ci/go-test-all.sh

test-race:
	GO_TEST_RACE_TIMEOUT=$(GO_TEST_RACE_TIMEOUT) ./scripts/ci/go-test-race.sh

test-race-ci:
	GO_TEST_RACE_TIMEOUT=$(GO_TEST_RACE_TIMEOUT) ./scripts/ci/go-test-race.sh

# --- Benchmarks ---

# Run microbenchmarks only (skip unit tests).
# Override BENCH and BENCH_COUNT as needed.
BENCH ?= .
BENCH_COUNT ?= 1

bench:
	go test ./... -run ^$$ -bench "$(BENCH)" -benchmem -count=$(BENCH_COUNT)

# Same as bench, but uses CI tag (can skip GUI-heavy code paths).
bench-ci:
	go test -tags=ci ./... -run ^$$ -bench "$(BENCH)" -benchmem -count=$(BENCH_COUNT)

# --- Hygiene ---

fmt:
	gofmt -w .

vet:
	go vet ./...
