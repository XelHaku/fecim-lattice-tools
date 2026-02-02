# Operations Runbook - FeCIM Lattice Tools

> **Purpose:** Deployment procedures, monitoring, common issues, and rollback procedures

## Build & Deploy

### Standard Build

```bash
# Build unified visualizer
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Or use launch script (builds and runs)
./launch.sh
```

### Release Build

```bash
# Optimized binary (smaller, faster)
go build -ldflags="-s -w" -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Cross-compile for Linux (from any OS)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o fecim-lattice-tools-linux ./cmd/fecim-lattice-tools

# Cross-compile for macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o fecim-lattice-tools-mac ./cmd/fecim-lattice-tools

# Cross-compile for macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o fecim-lattice-tools-mac-arm ./cmd/fecim-lattice-tools

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o fecim-lattice-tools.exe ./cmd/fecim-lattice-tools
```

### Pre-Deployment Checklist

```bash
# 1. Run all tests (see CI for current count)
go test ./... && echo "PASS"

# 2. Run with race detector
go test -race ./... && echo "NO RACES"

# 3. Check for vet issues
go vet ./... && echo "VET OK"

# 4. Format code
go fmt ./...

# 5. Build release binary
go build -ldflags="-s -w" -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# 6. Verify binary runs
./fecim-lattice-tools --help
```

### Build All Standalone Modules

```bash
# Build all 7 demo binaries
./scripts/build-all.sh

# Output:
# module1-hysteresis/hysteresis
# module2-crossbar/crossbar-gui
# module3-mnist/mnist-gui
# module4-circuits/circuits-gui
# module5-comparison/comparison-gui
```

## Monitoring & Diagnostics

### Logging Levels

| Level | Flag | Use Case |
|-------|------|----------|
| Off | `--verbosity 0` | Production |
| Info | `--verbosity 1` | Normal operation |
| Debug | `--verbosity 2` | Troubleshooting |
| Trace | `--verbosity 3` | Deep debugging |

```bash
./fecim-lattice-tools --verbosity 2  # Debug mode
```

**Log files are written to:** `logs/` directory with datetime stamps.

### Runtime Profiling

Enable pprof for production debugging:

```go
// Add to main.go temporarily
import _ "net/http/pprof"
import "net/http"

func init() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()
}
```

Then access:
- CPU: `http://localhost:6060/debug/pprof/profile?seconds=30`
- Memory: `http://localhost:6060/debug/pprof/heap`
- Goroutines: `http://localhost:6060/debug/pprof/goroutine`

```bash
# Analyze CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze memory
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Health Checks

**Manual verification:**
```bash
# Check binary starts
./fecim-lattice-tools &
PID=$!
sleep 3
kill -0 $PID && echo "Running OK" || echo "Failed to start"
kill $PID
```

## Common Issues & Fixes

### Issue: Black Screen / No Rendering

**Symptoms:** Application starts but window is black or blank

**Causes:**
1. Missing OpenGL drivers
2. Wayland compatibility issue
3. GPU driver crash

**Fixes:**
```bash
# Try software rendering
FYNE_NO_GL=1 ./fecim-lattice-tools

# Check OpenGL support
glxinfo | grep "OpenGL version"

# Update GPU drivers (Ubuntu)
sudo ubuntu-drivers autoinstall
```

### Issue: Window Resize Loop (Wayland/Sway)

**Symptoms:** Window constantly resizing, CPU spike

**Cause:** MinSize feedback loop with Wayland tiling WM

**Fix:** The app uses `ForceMinSizeLayout` wrapper in `cmd/fecim-lattice-tools/main.go:52`. If issue persists:
```bash
# Run under X11 instead
GDK_BACKEND=x11 ./fecim-lattice-tools
```

### Issue: CGO/GCC Build Errors

**Symptoms:** Build fails with gcc errors

**Fixes:**
```bash
# Linux (Ubuntu/Debian) - Install build dependencies
sudo apt-get install -y gcc libgl1-mesa-dev libx11-dev \
  libxinerama-dev libxrandr-dev libxcursor-dev libxi-dev libxext-dev libxfixes-dev

# Linux (Fedora/RHEL)
sudo dnf install -y gcc mesa-libGL-devel libX11-devel libXcursor-devel \
  libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel

# macOS - Install Xcode tools
xcode-select --install

# Verify gcc works
gcc --version
```

### Issue: Race Condition Panic

**Symptoms:** Random panics, especially during tab switches

**Cause:** UI updated from non-main goroutine

**Fix:** Wrap ALL UI updates in `fyne.Do()`:
```go
// Find the offending code
go build -race -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
# Race detector will print stack trace

// Fix: wrap in fyne.Do()
fyne.Do(func() {
    widget.SetText("value")
})
```

### Issue: Recording Not Working

**Symptoms:** Record button does nothing or produces empty file

**Fixes:**
```bash
# Check ffmpeg installed
which ffmpeg || sudo apt-get install ffmpeg

# Check permissions
mkdir -p recordings
chmod 755 recordings

# Verify ffmpeg works
ffmpeg -version
```

### Issue: Fonts Look Wrong / Blurry

**Symptoms:** Text rendering issues on HiDPI displays

**Fixes:**
```bash
# Set DPI scaling (Fyne usually auto-detects)
export FYNE_SCALE=1.5

# Or force standard DPI
export GDK_DPI_SCALE=1
```

### Issue: Module Fails to Initialize

**Symptoms:** Specific tab crashes on load

**Diagnosis:**
```bash
# Run with debug logging
./fecim-lattice-tools --verbosity 3 2>&1 | tee debug.log

# Look for initialization errors
grep -i "error\|panic\|fail" debug.log
```

**Common causes:**
1. Missing embedded interface implementation (`BuildContent()`, `Start()`, `Stop()`)
2. Nil pointer in BuildContent()
3. Resource loading failure

### Issue: High Memory Usage

**Symptoms:** Memory grows over time

**Diagnosis:**
```bash
# Monitor memory
watch -n 1 'ps -o rss= -p $(pgrep fecim)'

# Profile with pprof
go tool pprof http://localhost:6060/debug/pprof/heap
```

**Common causes:**
1. Goroutine leak (check demo Stop() implementations)
2. History buffer growing unbounded
3. Image cache not cleared

### Issue: Tests Fail in CI

**Symptoms:** Tests pass locally but fail in CI

**Cause:** GUI tests require display server

**Fix:** Skip GUI tests in headless:
```go
func TestGUIComponent(t *testing.T) {
    if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
        t.Skip("Skipping GUI test in headless environment")
    }
    // ... test code
}
```

### Issue: MNIST Data Not Found

**Symptoms:** Module 3 fails to load training data

**Fix:**
```bash
# Download MNIST data
cd module3-mnist/data
wget -q http://yann.lecun.com/exdb/mnist/train-images-idx3-ubyte.gz
wget -q http://yann.lecun.com/exdb/mnist/train-labels-idx1-ubyte.gz
wget -q http://yann.lecun.com/exdb/mnist/t10k-images-idx3-ubyte.gz
wget -q http://yann.lecun.com/exdb/mnist/t10k-labels-idx1-ubyte.gz

# Or use the training script (auto-downloads)
./module3-mnist/scripts/train_all_sizes.sh
```

### Issue: Module 6 EDA Tools Not Working

**Symptoms:** Yosys/OpenROAD commands fail

**Fixes:**
```bash
# Install Graphviz for schematic visualization
sudo apt install graphviz

# For full OpenLane flow, use Docker
docker pull efabless/openlane:latest

# Verify Yosys (if installed)
yosys --version
```

## Rollback Procedures

### Quick Rollback (Git)

```bash
# List recent commits
git log --oneline -10

# Revert to previous working commit
git checkout <commit-hash>
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Or revert last commit (creates new commit)
git revert HEAD
```

### Dependency Rollback

```bash
# Check current dependency versions
go list -m all | grep fyne

# Downgrade specific dependency
go get fyne.io/fyne/v2@v2.7.0

# Rebuild
go build ./cmd/fecim-lattice-tools
```

### Full Environment Reset

```bash
# Clear Go module cache
go clean -modcache

# Re-download dependencies
go mod download

# Rebuild from scratch
go build -a ./cmd/fecim-lattice-tools
```

## Performance Tuning

### Benchmark Specific Functions

```bash
# Run benchmarks
go test -bench=. ./module2-crossbar/pkg/crossbar

# With memory allocation stats
go test -bench=. -benchmem ./module2-crossbar/pkg/crossbar
```

### Optimize Hot Paths

**Known performance-critical areas:**
1. `crossbar.MVM()` - Matrix-vector multiply
2. `preisach.Calculate()` - Hysteresis computation
3. GUI refresh loops (should be ~50ms frame time)

**Profile before optimizing:**
```bash
go test -cpuprofile=cpu.prof -bench=BenchmarkMVM ./module2-crossbar/pkg/crossbar
go tool pprof cpu.prof
```

### MNIST Benchmarking

```bash
# Run comparison benchmark against literature
./module3-mnist/scripts/benchmark.sh

# Expected results:
# Float32 baseline:           98.1%
# 30-level demo baseline quant, no noise:   96.8%
# 30-level quant, noise=0.08: 87.0% (matches conference claim; unverified)
```

## Automated Git Commits

For long development sessions:

```bash
# One-time commit in 12 hours
./commit-push.sh -12

# Hourly commits (background)
./commit-push.sh --periodically

# Stop hourly commits
./commit-push.sh --stop
```

## Security Notes

### No Secrets Required

This application:
- Has no network connectivity (offline tool)
- Has no authentication
- Has no sensitive data storage
- No `.env` file needed

### File Permissions

Ensure output directories are writable:
```bash
mkdir -p screenshots recordings output logs
chmod 755 screenshots recordings output logs
```

## Contact & Escalation

| Issue Type | Resource |
|------------|----------|
| Build errors | `docs/development/WORKFLOWS.md#troubleshooting` |
| Physics questions | `docs/cim/HONESTY_AUDIT.md` |
| GUI issues | `docs/development/GUI/FYNE_NOTES.md` |
| Testing | `docs/development/TESTING.md` |
| General development | `docs/CONTRIB.md` |
| Project rules | `CLAUDE.md` |

---

**Last Updated:** 2026-01-29 | **Go Version:** 1.24+ (toolchain go1.24.12) | **Fyne Version:** 2.7.2
