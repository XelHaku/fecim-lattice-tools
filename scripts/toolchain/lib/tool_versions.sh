#!/usr/bin/env bash

# Helpers for comparing locally installed tool versions against documented pins.

get_current_version() {
  local tool="$1"
  case "$tool" in
    "ngspice")
      command -v ngspice >/dev/null 2>&1 || { echo ""; return; }
      ngspice -v 2>&1 | sed -n 's/.*ngspice-\([0-9][0-9]*\).*/\1/p' | head -n1
      ;;
    "iverilog")
      command -v iverilog >/dev/null 2>&1 || { echo ""; return; }
      iverilog -V 2>&1 | sed -n 's/.*version \([0-9][0-9.]*\).*/\1/p' | head -n1
      ;;
    "verilator")
      command -v verilator >/dev/null 2>&1 || { echo ""; return; }
      verilator --version 2>&1 | awk '{print $2}' | head -n1
      ;;
    "go")
      command -v go >/dev/null 2>&1 || { echo ""; return; }
      go version | awk '{print $3}' | sed 's/^go//' | head -n1
      ;;
    "python")
      command -v python3 >/dev/null 2>&1 || { echo ""; return; }
      python3 --version 2>&1 | awk '{print $2}' | head -n1
      ;;
    *)
      echo ""
      ;;
  esac
}

extract_pin() {
  local key="$1"
  local pin_file="$2"
  case "$key" in
    ngspice)
      grep -E '^\| ngspice \|' "$pin_file" | awk -F'|' '{gsub(/`/,"",$3); gsub(/ /,"",$3); print $3}'
      ;;
    iverilog)
      grep -E '^\| Icarus Verilog .*\|' "$pin_file" | awk -F'|' '{gsub(/`/,"",$3); gsub(/ /,"",$3); print $3}'
      ;;
    verilator)
      grep -E '^\| Verilator \|' "$pin_file" | awk -F'|' '{gsub(/`/,"",$3); gsub(/ /,"",$3); print $3}'
      ;;
    go)
      grep -E '^\| Go toolchain \|' "$pin_file" | awk -F'|' '{gsub(/`/,"",$3); gsub(/ /,"",$3); gsub(/\.x$/,"",$3); gsub(/x$/,"",$3); print $3}'
      ;;
    python)
      grep -E '^\| Python scientific stack .*\|' "$pin_file" | awk -F'|' '{gsub(/`/,"",$3); gsub(/ /,"",$3); split($3,a,","); sub(/^numpy==/,"",a[1]); print a[1]}'
      ;;
  esac
}
