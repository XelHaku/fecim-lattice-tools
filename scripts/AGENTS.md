<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# scripts/ — Build, CI, and Development Automation

**Purpose:** Build scripts, CI/CD pipelines, validation runners, and development toolchain automation. Centralizes testing, validation, and build orchestration.

**Status:** Production
**Stability:** High (mature automation)
**Coverage:** Covers Go builds, tests, race detection, validation, and screenshot generation

## Key Files

| File | Purpose | Key Operations |
|------|---------|-----------------|
| `run_full_validation.sh` | Full validation suite runner | Physics regression, config validation, integration tests |
| `reproduce_validation.sh` | One-command validation reproduction | For CI and reproducibility verification |
| `run_headless_ispp_regressions.sh` | Headless ISPP regression tests | 9 materials × 2 engines |
| `analyze_validation.sh` | Post-validation analysis and reporting | Parse test results, generate summary |
| `screenshot_all.sh` | Generate screenshots of all modules | GUI documentation and visual validation |
| `lk_perf_evidence.sh` | Performance evidence collection for L-K solver | Benchmarking and timing |
| `lk_log_stats.py` | Extract and analyze L-K solver logs | Performance metrics post-processing |
| `calib-guard.sh` | Calibration guard checks | Pre-commit validation |
| `pre-commit-calib-guard.sh` | Git pre-commit hook | Prevents invalid calibration commits |
| `submodules-setup.sh` | Initialize external tool submodules | Heracles, CrossSim, etc. |

## Subdirectories

| Directory | Purpose | Key Scripts |
|-----------|---------|-------------|
| `ci/` | CI/CD pipeline definitions and helpers | GitHub Actions, build scripts |
| `toolchain/` | Development toolchain: linters, formatters, language servers | Go tools setup |

## For AI Agents

### Working in This Directory

**When adding a build script:**
1. Use bash (`.sh`), Python (`.py`), or Go
2. Include header comment with purpose and usage examples
3. Add error handling and exit code validation
4. Test locally before committing
5. Document in this AGENTS.md file

**When adding a validation script:**
1. Call `go test ./...` with appropriate flags (see patterns below)
2. Parse and summarize results
3. Generate exit code 0 (success) or 1 (failure) for CI
4. Log to stdout and optional summary file

**When adding CI/CD jobs:**
1. Put workflow definitions in `ci/` directory
2. Use GitHub Actions YAML format
3. Reference build scripts from root `scripts/` directory
4. Test workflow locally with `act` tool if possible
5. Document in `ci/README.md`

**When running validation:**
1. Use `run_full_validation.sh` for comprehensive suite
2. Use `reproduce_validation.sh` for CI reproducibility
3. Check exit code: 0 = pass, 1 = fail
4. Review generated reports in output directory

### Testing Requirements

**All scripts must handle errors gracefully:**
```bash
set -e              # Exit on error
set -u              # Exit on undefined variable
trap 'cleanup' EXIT # Cleanup on exit
```

**Validation scripts must:**
1. Return exit code 0 (success) or 1 (failure)
2. Generate parseable output (JSON, CSV, or structured text)
3. Log test names and results
4. Summarize pass/fail counts

**Build scripts must:**
1. Verify dependencies before use
2. Clean previous artifacts (optional cleanup)
3. Compile successfully (exit 0)
4. Generate binary with correct permissions

**Performance scripts must:**
1. Warm up (discard first run)
2. Run multiple iterations (N=10 typical)
3. Report mean, min, max, std deviation
4. Include system info (Go version, GOOS, GOARCH)

### Common Patterns

**Run all Go tests with race detector:**
```bash
go test -race ./...
```
Found in `ci/go-test-race.sh`

**Run specific test suite:**
```bash
go test -v ./validation/ -run TestPhysicsRegression
```
Found in `run_full_validation.sh`

**Generate golden data (with guard):**
```bash
FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./validation/...
```
Guard checks in `calib-guard.sh` prevent accidental regeneration.

**Headless ISPP test matrix:**
```bash
# Tests 9 materials × 2 engines (LK-based vs. level-based)
go test -v ./cmd/fecim-lattice-tools/ -run TestISPPEngineMatrix
```
Found in `run_headless_ispp_regressions.sh`

**Screenshot generation:**
```bash
# Launches GUI in headless mode, captures screenshots
./fecim-lattice-tools -mode=screenshot -module=module1
```
Found in `screenshot_all.sh`

**Performance profiling:**
```bash
go test -cpuprofile=cpu.prof -benchmem -bench=Landau ./validation/...
go tool pprof cpu.prof
```
Used by `lk_perf_evidence.sh`

## Dependencies

### Internal
- `cmd/fecim-lattice-tools/` — Main application
- `validation/` — Test suites
- All modules — Tested by CI scripts

### External
- `bash` — Script interpreter (version 4.0+)
- `go` — Go compiler and test runner
- `python3` — For data analysis (lk_log_stats.py)
- Optional: `jq` (JSON querying), `git` (version control)

## MANUAL

**Running Full Validation Locally:**
```bash
cd <repo-root>
./scripts/run_full_validation.sh
```
Runs: physics regression, config validation, integration tests, determinism checks.

**One-Command Reproducibility Check:**
```bash
./scripts/reproduce_validation.sh
```
Verifies entire validation suite runs consistently. Used in CI to detect flaky tests.

**ISPP Regression Testing:**
```bash
./scripts/run_headless_ispp_regressions.sh
```
Tests 9 materials (HZO, AlScN, cryogenic variants) with both LK-based and level-based ISPP engines.
Exit code: 0 if all pass, 1 if any fail.

**Performance Benchmarking:**
```bash
./scripts/lk_perf_evidence.sh
```
Collects L-K solver performance data, generates timing report.
Results: `output/lk_perf_evidence.csv`

**Pre-Commit Calibration Guard:**
```bash
./scripts/pre-commit-calib-guard.sh
```
Install as `.git/hooks/pre-commit`:
```bash
cp scripts/pre-commit-calib-guard.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```
Prevents commits with invalid calibration files.

**Generate Documentation Screenshots:**
```bash
./scripts/screenshot_all.sh
```
Captures GUI screenshots for all modules.
Output: `output/screenshots/module*.png`

**Analysis of Validation Results:**
```bash
./scripts/analyze_validation.sh
```
Parses test logs, generates summary report.
Output: `output/validation_summary.txt`

**Setting Up External Tool Submodules:**
```bash
./scripts/submodules-setup.sh
```
Clones Heracles, CrossSim, and other optional external tools.
(Optional; tools are only needed for cross-validation)

**Build Without Tests:**
```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```
Fast build, no validation. For development only.

**Build With Full CI Checks:**
```bash
./scripts/ci/go-test-all.sh
```
Runs: go vet, go fmt, go test, race detector.
Equivalent to CI pipeline checks.

**Python Data Analysis:**
```bash
python3 scripts/lk_log_stats.py < output/lk_solver.log
```
Extracts timing, convergence, and performance statistics from solver logs.

**Exit Codes:**
- `0` — All checks passed
- `1` — One or more checks failed (see stdout for details)
- `2` — Missing dependencies or setup error

**Debugging Failed Tests:**
```bash
# Run single test with verbose output
go test -v -run TestName ./package/...

# Run test with logging
FECIM_LOG_LEVEL=debug go test -v ./package/...

# Run with race detector
go test -race -run TestName ./package/...
```

**CI Pipeline Trigger:**
Pushing to `main` or opening PR automatically runs:
1. `scripts/ci/go-test-all.sh` — Basic build and tests
2. `scripts/run_full_validation.sh` — Full validation suite
3. `scripts/run_headless_ispp_regressions.sh` — Physics regression matrix

Check GitHub Actions tab for results.

**Adding New Validation:**
1. Create script in `scripts/` (e.g., `new_check.sh`)
2. Make executable: `chmod +x scripts/new_check.sh`
3. Test locally: `./scripts/new_check.sh`
4. Add to `run_full_validation.sh` if part of main suite
5. Document in this AGENTS.md file

**Performance Regression Detection:**
Benchmarks are tracked in CI:
```bash
go test -bench=. -benchmem ./validation/benchmarks/... | tee benchmarks/results.txt
# Commit results: git add benchmarks/results.txt
```
CI compares new results to baseline for regressions.

**Headless Mode for Docker/CI:**
All scripts work in headless environments (no X11 needed):
```bash
# Works in Docker
docker run --rm golang:1.25 /bin/bash scripts/run_full_validation.sh
```
