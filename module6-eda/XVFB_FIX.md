# X11/Display Fix for KLayout and OpenROAD Image Generation

## Problem
KLayout and OpenROAD image generation commands (`save_image`) require GUI/X11 display access, which fails in Docker containers without proper X11 forwarding:

- **KLayout**: "Authorization required, but no authorization protocol specified" and "Could not connect to any X display"
- **OpenROAD**: "exit status 1" when calling `save_image` TCL command

## Solution
Implemented Xvfb (X Virtual FrameBuffer) wrapper for headless rendering in Docker mode.

### Changes Made

#### 1. `/module6-eda/pkg/openlane/runner.go`

**runDockerKLayout()** - Modified to use Xvfb wrapper:
```go
// Before: Direct klayout entrypoint with X11 passthrough
args := []string{"run", "--rm", "--entrypoint", "klayout", ...}

// After: sh entrypoint with Xvfb wrapper
args := []string{"run", "--rm", "--entrypoint", "sh", ...}
xvfbCmd := "Xvfb :99 -screen 0 1024x768x24 -nolisten tcp > /dev/null 2>&1 & sleep 1 && export DISPLAY=:99 && klayout ..."
```

**runDockerOpenROAD()** - Modified to use Xvfb wrapper:
```go
// Before: Direct openroad entrypoint with X11 passthrough
args := []string{"run", "--rm", "--entrypoint", "openroad", ...}

// After: sh entrypoint with Xvfb wrapper
args := []string{"run", "--rm", "--entrypoint", "sh", ...}
xvfbCmd := "Xvfb :99 -screen 0 1024x768x24 -nolisten tcp > /dev/null 2>&1 & sleep 1 && export DISPLAY=:99 && openroad ..."
```

### How It Works

1. **Xvfb :99** - Creates virtual X display on display number 99
2. **-screen 0 1024x768x24** - Configures screen size and color depth
3. **-nolisten tcp** - Security: disable network listening
4. **> /dev/null 2>&1 &** - Background the Xvfb process, suppress output
5. **sleep 1** - Wait for Xvfb to initialize
6. **export DISPLAY=:99** - Point X clients to virtual display
7. **klayout/openroad ...** - Run the actual command with GUI access

### Benefits

- ✅ No user X11 configuration required
- ✅ Works in headless environments (CI/CD, servers)
- ✅ No security risks from X11 socket forwarding
- ✅ Compatible with OpenLane Docker image (includes Xvfb)
- ✅ Native mode unchanged (still uses host display if available)

### Testing

All existing tests pass:
```bash
go test ./module6-eda/...
# ok  	fecim-lattice-tools/module6-eda/pkg/validation	0.010s
```

### Usage

No changes required for users. The fix is transparent:

```go
// Generate layout image (automatically uses Xvfb in Docker mode)
result, err := validation.GenerateLayoutImage(defPath, lefPath, outputPath, manager, config)

// Generate OpenROAD image (automatically uses Xvfb in Docker mode)
result, err := validation.GenerateOpenROADImage(defPath, lefPath, outputPath, manager, config)
```

### Requirements

- OpenLane Docker image (efabless/openlane:latest) includes Xvfb
- No additional dependencies for users

### Backwards Compatibility

- **Docker mode**: Now uses Xvfb (fixed)
- **Native mode**: Unchanged (uses host DISPLAY)
- All existing code continues to work without modification
