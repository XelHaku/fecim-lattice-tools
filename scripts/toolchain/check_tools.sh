#!/usr/bin/env bash
set -u

# Toolchain checker for FeCIM external tooling.
# Exits 1 only when required tools are missing.

required_missing=0

check_tool() {
  local label="$1"
  local cmd="$2"
  local version_cmd="$3"
  local required="$4" # required|optional

  if command -v "$cmd" >/dev/null 2>&1; then
    local version
    version=$(eval "$version_cmd" 2>&1 | head -n 1)
    if [[ -z "$version" ]]; then
      version="(version string unavailable)"
    fi
    printf "%-20s FOUND    %s\n" "$label" "$version"
  else
    if [[ "$required" == "required" ]]; then
      printf "%-20s NOT FOUND (required)\n" "$label"
      required_missing=1
    else
      printf "%-20s NOT FOUND (optional)\n" "$label"
    fi
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

if command -v python3 >/dev/null 2>&1; then
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
