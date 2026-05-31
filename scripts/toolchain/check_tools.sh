#!/usr/bin/env bash
set -u

# Toolchain checker for FeCIM external tooling.
# Exits 1 only when required tools are missing.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/toolchain_status.sh
source "$SCRIPT_DIR/lib/toolchain_status.sh"

required_missing=0

check_tool() {
  if ! print_tool_status "$@"; then
    required_missing=1
  fi
}

check_tool "go" "go" "go version" "required"
check_tool "ngspice" "ngspice" "ngspice -v" "required"
check_tool "heracles" "heracles" "heracles --version" "optional"
check_tool "crosssim" "crosssim" "crosssim --version" "optional"
check_tool "iverilog" "iverilog" "iverilog -V" "optional"
check_tool "verilator" "verilator" "verilator --version" "optional"
check_tool "openroad" "openroad" "openroad -version" "optional"
check_tool "openlane" "openlane" "openlane --version" "optional"
check_tool "python3" "python3" "python3 --version" "optional"

if command_exists python3; then
  py_stack=$(python3 - <<'PY'
try:
    import numpy, scipy
    print(f"numpy {numpy.__version__}; scipy {scipy.__version__}")
except Exception as e:
    print(f"python stack unavailable: {e}")
PY
)
  printf "%-20s FOUND    %s\n" "numpy/scipy" "$py_stack"
else
  printf "%-20s NOT FOUND (optional)\n" "numpy/scipy"
fi

exit $required_missing
