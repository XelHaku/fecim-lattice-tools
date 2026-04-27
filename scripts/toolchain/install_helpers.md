# External Tool Installation Guide

Reproducible install steps for FeCIM external validation tools.

## Required Tools

### Go (1.25+)
```bash
# Linux (via official installer)
curl -LO https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# macOS (via Homebrew)
brew install go
```

### ngspice (42+)
```bash
# Ubuntu/Debian
sudo apt-get install -y ngspice

# macOS
brew install ngspice

# Verify
ngspice --version
```

## Optional Tools

### Iverilog (12.0+)
```bash
# Ubuntu/Debian
sudo apt-get install -y iverilog

# macOS
brew install icarus-verilog

# Verify
iverilog -V
```

### Verilator (5.0+)
```bash
# Ubuntu/Debian
sudo apt-get install -y verilator

# macOS
brew install verilator

# Verify
verilator --version
```

### Python Scientific Stack (numpy/scipy)
```bash
pip3 install numpy scipy

# Verify
python3 -c "import numpy; import scipy; print('OK')"
```

### OpenROAD / OpenLane2
```bash
# Follow official OpenLane2 installation:
# https://openlane2.readthedocs.io/en/latest/getting_started/installation/index.html
# Requires Docker or Nix

# Verify
openlane --version
```

### Heracles
```bash
# Academic tool from University of Groningen
# Request access: https://www.rug.nl/research/zernike/micromechanics/
# No public package manager install available
```

### CrossSim (Sandia)
```bash
# Clone from GitHub:
git clone https://github.com/sandialabs/cross-sim.git
cd cross-sim && pip3 install -e .

# Verify
python3 -c "import cross_sim; print('OK')"
```

## After Installation

Run the tool checker to verify:
```bash
bash scripts/toolchain/check_tools.sh
```

All required tools must show `FOUND`. Optional tools show `NOT FOUND (optional)` if absent.
