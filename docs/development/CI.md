# Continuous Integration (CI)

CI is implemented with **GitHub Actions**: `.github/workflows/ci.yml`.

## What CI runs

### macOS + Ubuntu

- `make test-ci`
  - Internally runs: `go test -tags=ci -count=1 -shuffle=off -trimpath -timeout ... ./...`
  - Script: `scripts/ci/go-test-all.sh`

### Windows

- `go test -tags=ci -count=1 -shuffle=off -trimpath -timeout 10m ./...`

### Race detector (Ubuntu, targeted packages)

- `make test-race-ci`
  - Script: `scripts/ci/go-test-race.sh`
  - Runs `go test -tags=ci -race` on a curated package list to keep runtime stable.

## The `ci` build tag

Tests in this repo may use the `ci` build tag to:

- avoid GUI-heavy code paths in headless runners
- disable slow/interactive behavior
- keep execution deterministic

If you are debugging a CI failure locally, prefer:

```bash
make test-ci
# or
GO_TEST_TIMEOUT=15m go test -tags=ci -count=1 -shuffle=off -trimpath -timeout 15m ./...
```

## Reproducing common CI settings locally

```bash
# All packages, deterministic ordering, no cached results
GO_TEST_TIMEOUT=10m go test -tags=ci -count=1 -shuffle=off -trimpath -timeout 10m ./...

# Targeted race set (mirrors CI)
GO_TEST_RACE_TIMEOUT=20m make test-race-ci
```

## Headless note (Linux)

CI is designed to be **headless**: unit tests should not require a display.

Some GUI/layout audit tests are intentionally skipped unless a display is present. If you need to run them on a server, use Xvfb:

```bash
sudo apt-get install -y xvfb
xvfb-run -a go test -v ./cmd/fecim-lattice-tools/... -run LayoutAudit
```

See `docs/development/HEADLESS.md`.
