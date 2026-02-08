# Security Audit - FeCIM Lattice Tools

**Audit Date:** 2026-02-07  
**Auditor:** OpenClaw Security Subagent  
**Scope:** File path handling, input validation, unsafe operations

---

## Executive Summary

This security audit identified several areas of concern in the FeCIM Lattice Tools codebase, primarily related to **command injection risks** in external tool execution and **path traversal vulnerabilities** in file handling. Most findings are **Medium** severity given the local desktop application context, but would be **Critical** if the application were to expose any network services.

| Severity | Count | Description |
|----------|-------|-------------|
| High     | 2     | Command injection in EDA tool execution |
| Medium   | 5     | Path handling and validation issues |
| Low      | 4     | File permissions and resource management |
| Info     | 3     | Best practice recommendations |

---

## High Severity Findings

### 1. Command Injection in Yosys Execution

**File:** `module6-eda/pkg/validate/yosys.go` (line 14)  
**File:** `module6-eda/pkg/validation/yosys.go` (lines 38, 109)

**Issue:** File paths are interpolated directly into Yosys command strings using `fmt.Sprintf` without sanitization:

```go
cmdStr := fmt.Sprintf("read_verilog %s; hierarchy -check; check", verilogPath)
cmd := exec.Command("yosys", "-p", cmdStr)
```

**Risk:** A malicious filename containing Yosys command separators (`;`, newlines) could execute arbitrary Yosys commands:
- Example: A file named `test; write_verilog /etc/passwd.v;.v` would inject commands

**Recommendation:**
- Validate that file paths contain only safe characters (alphanumeric, `-`, `_`, `.`, `/`)
- Use absolute paths and verify the file exists before building commands
- Consider using Yosys script files instead of inline commands

```go
// Example fix
func sanitizePath(path string) error {
    if strings.ContainsAny(path, ";\n\r`$\"'|&") {
        return fmt.Errorf("invalid characters in path: %s", path)
    }
    return nil
}
```

### 2. Command Injection in OpenROAD/KLayout Docker Commands

**File:** `module6-eda/pkg/openlane/runner.go` (lines 78, 187-189)

**Issue:** File paths and script names are interpolated into shell commands passed to Docker:

```go
xvfbCmd := fmt.Sprintf("Xvfb :99 ... && openroad -no_splash -exit /design/%s", scriptName)
```

**Risk:** If `scriptName` or paths contain shell metacharacters, arbitrary commands could execute inside the Docker container.

**Recommendation:**
- Sanitize all path components before shell interpolation
- Use `filepath.Clean()` and validate against directory traversal
- Consider using `exec.Command` with separate arguments instead of shell strings

---

## Medium Severity Findings

### 3. Path Traversal in File Discovery

**File:** `shared/utils/path_discovery.go`

**Issue:** The `FindDirectory`, `FindFile`, and `FindModuleDataDir` functions search relative paths without sanitizing input:

```go
candidates := []string{
    dirName,
    filepath.Join("..", dirName),
    filepath.Join("..", "..", dirName),
}
```

**Risk:** While these functions don't accept external input directly, if used with user-controlled strings, they could access files outside the intended directory.

**Recommendation:**
- Always use `filepath.Clean()` on paths before use
- Validate that resolved paths remain within expected directories:

```go
func isWithinDir(targetPath, baseDir string) bool {
    rel, err := filepath.Rel(baseDir, targetPath)
    if err != nil {
        return false
    }
    return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
```

### 4. Unsafe File Path Construction in Screenshot/Recording

**File:** `cmd/fecim-lattice-tools/main.go` (lines 122-129, 159-165)

**Issue:** Screenshot and recording filenames include `sectionName` without full sanitization:

```go
filename := filepath.Join(screenshotDir, fmt.Sprintf("fecim_%s_%s.png", sectionName, timestamp))
```

**Risk:** While `sectionName` comes from internal `viewNames` array (safe), the pattern could be problematic if extended.

**Recommendation:**
- Add explicit sanitization for any path components derived from user input
- Document the security assumption that `sectionName` is trusted

### 5. Python Script Execution with Embedded Code

**File:** `shared/validation/crossbar_tools.go` (lines 153-160, 225-250, 276-306)

**Issue:** Python scripts are constructed with module names:

```go
script := fmt.Sprintf(`import %s; version = getattr(%s, '__version__', 'unknown')`, moduleName, moduleName)
cmd := exec.CommandContext(ctx, pythonCmd, "-c", script)
```

**Risk:** The `moduleName` is hardcoded ("crosssim", "badcrossbar") so this is currently safe. However, if extended to user-supplied module names, this would allow arbitrary Python code execution.

**Recommendation:**
- Add a whitelist check for allowed module names
- Document that module names must not come from external input

### 6. Insufficient Input Validation in GUI Controls

**File:** `module4-circuits/pkg/gui/tab_unified.go` (lines 753-820)  
**File:** `module6-eda/pkg/gui/tabs/builder_validation_tab.go` (lines 64-84, 366-382)

**Issue:** Numeric inputs from GUI entries are parsed but validation is minimal:

```go
voltage, err := strconv.ParseFloat(voltageStr, 64)
if err != nil {
    return // Just returns on error
}
```

**Risk:** While GUI context limits attack surface, extreme values could cause:
- Integer overflow/underflow
- Resource exhaustion (very large array dimensions)
- Floating-point issues (Inf, NaN)

**Recommendation:**
- Add range validation for all numeric inputs
- Validate dimensions against sane maximums (e.g., max 10,000 for array size)

```go
const (
    MaxArrayDimension = 10000
    MinVoltage = -100.0
    MaxVoltage = 100.0
)
```

### 7. Environment Variable Trust

**File:** `module6-eda/pkg/openlane/config.go` (lines 36-42, 54)

**Issue:** Environment variables like `PDK_ROOT` and `HOME` are used to construct file paths without validation:

```go
pdkRoot := os.Getenv("PDK_ROOT")
if pdkRoot == "" {
    pdkRoot = filepath.Join(os.Getenv("HOME"), ".volare")
}
```

**Risk:** Maliciously set environment variables could redirect file operations.

**Recommendation:**
- Validate environment-derived paths exist and are directories
- Use `filepath.Clean()` on all environment-derived paths

---

## Low Severity Findings

### 8. File Permission Consistency

**Files:** Multiple (see grep for `0644`, `0755`)

**Observation:** File permissions are consistently used:
- `0644` for data files (read/write for owner, read for others)
- `0755` for directories

**Status:** ✅ Good practice followed

### 9. External Command Timeout Handling

**File:** `shared/validation/crossbar_tools.go`  
**File:** `module6-eda/pkg/openlane/runner.go`

**Observation:** Timeouts are implemented for external commands:
- Python module checks: 5-30 seconds
- OpenLane operations: 5-15 minutes
- Package installation: 120 seconds

**Status:** ✅ Timeouts in place, values are reasonable

**Minor Concern:** Some FFmpeg operations in `shared/recording/` don't have explicit timeouts - they rely on user stopping the recording.

### 10. Process Cleanup on Recording Stop

**File:** `cmd/fecim-lattice-tools/main.go` (lines 320-350)

**Issue:** Recording stop has a 2-second timeout for FFmpeg:

```go
case <-time.After(2 * time.Second):
    if cmd.Process != nil {
        _ = cmd.Process.Kill()
    }
```

**Status:** ✅ Good practice - process is killed if it doesn't stop gracefully

### 11. JSON Parsing Without Size Limits

**Files:** Multiple (`json.Unmarshal`, `json.NewDecoder`)

**Issue:** JSON parsing doesn't limit input size, which could cause memory exhaustion with very large files.

**Risk:** Low - input comes from local files in a desktop application context

**Recommendation:** For any future network-facing features, add size limits:
```go
decoder := json.NewDecoder(io.LimitReader(reader, maxSize))
```

---

## Informational Findings

### 12. Shell Script Security (commit-push.sh)

**File:** `commit-push.sh`

**Observations:**
- ✅ Uses `"${BASH_SOURCE[0]}"` safely with quotes
- ✅ PID file handling is reasonable
- ⚠️ Uses `systemd-run` which requires user session

**Status:** Acceptable for development automation script

### 13. No Network Services

**Observation:** The application does not expose any network services (HTTP, WebSocket, etc.). All external communication is:
- FFmpeg for screen/audio recording (local)
- Docker for EDA tools (local)
- Git for version control (user-initiated)

**Status:** ✅ Significantly reduces attack surface

### 14. Symlink Handling

**Observation:** No symlink-related code found (`Lstat`, `EvalSymlinks`, etc.)

**Recommendation:** If file operations may encounter symlinks, consider:
- Using `filepath.EvalSymlinks()` to resolve them
- Checking that resolved paths stay within allowed directories

---

## Recommendations Summary

### Priority 1 (Address Soon)
1. **Sanitize file paths** before interpolating into Yosys/OpenROAD commands
2. **Add character validation** for paths used in shell command strings
3. **Create a `SafePath` helper** that validates and cleans paths

### Priority 2 (Good Practice)
4. Add range validation for numeric GUI inputs
5. Validate environment-derived paths
6. Document security assumptions in code comments

### Priority 3 (Future Hardening)
7. Add JSON size limits if network features are added
8. Consider sandboxing external tool execution
9. Add symlink handling if needed

---

## Secure Coding Guidelines for Contributors

### Path Handling
```go
// Always clean and validate paths
cleanPath := filepath.Clean(userPath)
absPath, err := filepath.Abs(cleanPath)
if err != nil {
    return err
}
// Verify path is within expected directory
if !strings.HasPrefix(absPath, allowedBase) {
    return fmt.Errorf("path outside allowed directory")
}
```

### Command Execution
```go
// Prefer separate arguments over shell strings
cmd := exec.Command("tool", "--file", filePath) // ✅ Good

// Avoid string interpolation in shell commands
cmd := exec.Command("sh", "-c", "tool --file " + filePath) // ❌ Bad
```

### Input Validation
```go
// Validate numeric inputs with ranges
value, err := strconv.ParseFloat(input, 64)
if err != nil || value < MinValue || value > MaxValue {
    return fmt.Errorf("value out of range [%f, %f]", MinValue, MaxValue)
}
```

---

## Conclusion

The FeCIM Lattice Tools codebase follows many good security practices (timeouts, proper file permissions, no network exposure). The main areas requiring attention are:

1. **Command injection risks** in EDA tool execution - needs path sanitization
2. **Path validation** - needs consistent application across the codebase

Given the local desktop application context, the overall security posture is **acceptable** with the recommended improvements. The findings would become critical if the application were to expose network services or process untrusted input files.
