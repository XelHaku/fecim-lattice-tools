package compute

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	vk "github.com/vulkan-go/vulkan"
)

// LoadSPIRV loads a SPIR-V shader file from disk, trying multiple search paths.
// Search order:
//  1. Absolute path as-is
//  2. Relative to working directory
//  3. Relative to executable
//  4. shared/compute/shaders/ directory
func LoadSPIRV(path string) ([]byte, error) {
	// Try paths in order
	searchPaths := []string{
		path, // As-is (absolute or relative to cwd)
	}

	// Add relative to working directory
	if cwd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(cwd, path))
	}

	// Add relative to executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		searchPaths = append(searchPaths, filepath.Join(exeDir, path))
	}

	// Add shared/compute/shaders/ directory
	if cwd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(cwd, "shared", "compute", "shaders", path))
	}

	// Try each path
	var lastErr error
	for _, tryPath := range searchPaths {
		data, err := os.ReadFile(tryPath)
		if err == nil {
			// Validate SPIR-V magic number (0x07230203)
			if len(data) < 4 {
				return nil, fmt.Errorf("invalid SPIR-V file %s: too small (%d bytes)", path, len(data))
			}
			magic := binary.LittleEndian.Uint32(data[0:4])
			if magic != 0x07230203 {
				return nil, fmt.Errorf("invalid SPIR-V file %s: bad magic number 0x%08x", path, magic)
			}
			return data, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("failed to load SPIR-V shader %s: %w (tried %d paths)", path, lastErr, len(searchPaths))
}

// CreateShaderModule creates a Vulkan shader module from SPIR-V bytecode.
// The SPIR-V code must be in little-endian format.
func CreateShaderModule(ctx *VulkanContext, spirvCode []byte) (vk.ShaderModule, error) {
	if ctx == nil {
		return nil, fmt.Errorf("VulkanContext is nil")
	}
	if ctx.device == nil {
		return nil, fmt.Errorf("VulkanContext.device is nil")
	}
	if len(spirvCode) == 0 {
		return nil, fmt.Errorf("SPIR-V code is empty")
	}
	if len(spirvCode)%4 != 0 {
		return nil, fmt.Errorf("SPIR-V code length must be multiple of 4 bytes, got %d", len(spirvCode))
	}

	// Convert byte slice to uint32 slice
	codeUint32 := make([]uint32, len(spirvCode)/4)
	for i := range codeUint32 {
		codeUint32[i] = binary.LittleEndian.Uint32(spirvCode[i*4:])
	}

	createInfo := vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len(spirvCode)),
		PCode:    codeUint32,
	}

	var shaderModule vk.ShaderModule
	result := vk.CreateShaderModule(ctx.device, &createInfo, nil, &shaderModule)
	if result != vk.Success {
		return nil, fmt.Errorf("failed to create shader module: VkResult=%d", result)
	}

	return shaderModule, nil
}
