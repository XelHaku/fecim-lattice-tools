#!/usr/bin/env bash

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

first_output_line() {
  eval "$1" 2>&1 | head -n 1
}

print_tool_status() {
  local label="$1"
  local cmd="$2"
  local version_cmd="$3"
  local required="$4" # required|optional

  if command_exists "$cmd"; then
    local version
    version=$(first_output_line "$version_cmd")
    if [[ -z "$version" ]]; then
      version="(version string unavailable)"
    fi
    printf "%-20s FOUND    %s\n" "$label" "$version"
    return 0
  fi

  if [[ "$required" == "required" ]]; then
    printf "%-20s NOT FOUND (required)\n" "$label"
    return 1
  fi

  printf "%-20s NOT FOUND (optional)\n" "$label"
  return 0
}
