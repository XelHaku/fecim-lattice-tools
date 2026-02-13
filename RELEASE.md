### Build Instructions

**Prerequisites (all platforms):**
- Go 1.24+ (https://go.dev/dl/)
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
# Run validation suite
bash scripts/reproduce_validation.sh

# Expected: 70 packages PASS, 0 FAIL
```

### Minimum Hardware
- 4 GB RAM (8 GB recommended for large arrays)
- OpenGL 2.0+ capable GPU (for Fyne rendering)
- 200 MB disk space
