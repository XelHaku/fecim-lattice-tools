### Build Instructions

**Prerequisites (all platforms):**
- Go 1.25+ (https://go.dev/dl/)
- Git

**Linux additional deps:**
```bash
sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
```

**macOS additional deps:**
- Xcode Command Line Tools: `xcode-select --install`

**Windows additional deps:**
- MinGW-w64 or MSYS2 with GCC
- TDM-GCC also works

### Build & Run
```bash
git clone https://github.com/your-org/fecim-lattice-tools.git
cd fecim-lattice-tools
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools/
./fecim-lattice-tools
```

### Package as Desktop App (optional)
```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
fyne package -os linux -name "FeCIM Lattice Tools" -appID com.fecim.latticetools
```

### Cross-compilation
```bash
# Install fyne-cross
go install github.com/fyne-io/fyne-cross@latest

# Linux
fyne-cross linux -arch amd64

# macOS (requires macOS SDK)
fyne-cross darwin -arch amd64

# Windows
fyne-cross windows -arch amd64
```

### Verify Installation
```bash
# Run full validation suite
go test ./...

# Expected: >=70 packages PASS, 0 FAIL
```

### Web (WASM) demo (experimental)
This repository includes an experimental WASM entrypoint:
- `cmd/fecim-web` (currently loads Module 7 documentation viewer)

Build + run locally:
```bash
# Build
GOOS=js GOARCH=wasm go build -o web/fecim.wasm ./cmd/fecim-web
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/

# Serve
python3 -m http.server --directory web 8080
# Then open: http://localhost:8080
```

### Minimum Hardware
- 4 GB RAM (8 GB recommended for large arrays)
- OpenGL 2.0+ capable GPU (for Fyne rendering)
- 200 MB disk space
