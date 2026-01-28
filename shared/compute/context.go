// Package compute provides headless Vulkan compute context for GPU-accelerated operations.
package compute

import (
	"fmt"

	vk "github.com/vulkan-go/vulkan"
)

// VulkanContext manages headless Vulkan compute resources.
type VulkanContext struct {
	instance       vk.Instance
	physicalDevice vk.PhysicalDevice
	device         vk.Device
	computeQueue   vk.Queue
	computeFamily  uint32
	memoryProps    vk.PhysicalDeviceMemoryProperties
	commandPool    vk.CommandPool
	available      bool // true if Vulkan initialized successfully
}

// NewVulkanContext creates and initializes a headless Vulkan compute context.
// Returns a context with available=false if Vulkan is not available, rather than returning an error.
func NewVulkanContext() (*VulkanContext, error) {
	ctx := &VulkanContext{
		available: false,
	}

	// Initialize Vulkan using default proc addr (headless mode)
	if err := vk.SetDefaultGetInstanceProcAddr(); err != nil {
		return ctx, nil // Not available, but not an error
	}

	if err := vk.Init(); err != nil {
		return ctx, nil // Not available, but not an error
	}

	// Create Vulkan instance (no surface, no display extensions)
	if err := ctx.createInstance(); err != nil {
		return ctx, nil // Not available, but not an error
	}

	// Pick physical device with compute capabilities
	if err := ctx.pickPhysicalDevice(); err != nil {
		ctx.Destroy()
		return ctx, nil // Not available, but not an error
	}

	// Create logical device with compute queue
	if err := ctx.createLogicalDevice(); err != nil {
		ctx.Destroy()
		return ctx, nil // Not available, but not an error
	}

	// Create command pool for compute operations
	if err := ctx.createCommandPool(); err != nil {
		ctx.Destroy()
		return ctx, nil // Not available, but not an error
	}

	// Successfully initialized
	ctx.available = true
	return ctx, nil
}

// IsAvailable returns true if GPU compute is available.
func (c *VulkanContext) IsAvailable() bool {
	return c.available
}

// createInstance creates a Vulkan instance without any display extensions.
func (c *VulkanContext) createInstance() error {
	appInfo := vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PApplicationName:   safeString("FeCIM Compute"),
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        safeString("FeCIM"),
		EngineVersion:      vk.MakeVersion(1, 0, 0),
		ApiVersion:         vk.ApiVersion11,
	}

	createInfo := vk.InstanceCreateInfo{
		SType:            vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo: &appInfo,
		// No extensions needed for headless compute
		EnabledExtensionCount:   0,
		PpEnabledExtensionNames: nil,
	}

	var instance vk.Instance
	if result := vk.CreateInstance(&createInfo, nil, &instance); result != vk.Success {
		return fmt.Errorf("vkCreateInstance failed: %d", result)
	}
	c.instance = instance

	vk.InitInstance(c.instance)
	return nil
}

// pickPhysicalDevice selects a GPU with compute capabilities.
func (c *VulkanContext) pickPhysicalDevice() error {
	var deviceCount uint32
	vk.EnumeratePhysicalDevices(c.instance, &deviceCount, nil)
	if deviceCount == 0 {
		return fmt.Errorf("no Vulkan-capable GPU found")
	}

	devices := make([]vk.PhysicalDevice, deviceCount)
	vk.EnumeratePhysicalDevices(c.instance, &deviceCount, devices)

	// Pick first device with compute queue family
	for _, device := range devices {
		if computeFamily, found := c.findComputeQueueFamily(device); found {
			c.physicalDevice = device
			c.computeFamily = computeFamily
			return nil
		}
	}

	return fmt.Errorf("no GPU with compute capabilities found")
}

// findComputeQueueFamily finds a queue family with compute support.
func (c *VulkanContext) findComputeQueueFamily(device vk.PhysicalDevice) (uint32, bool) {
	var queueFamilyCount uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, nil)
	queueFamilies := make([]vk.QueueFamilyProperties, queueFamilyCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, queueFamilies)

	for i, qf := range queueFamilies {
		qf.Deref()
		if qf.QueueFlags&vk.QueueFlags(vk.QueueComputeBit) != 0 {
			return uint32(i), true
		}
	}

	return 0, false
}

// createLogicalDevice creates a logical device with compute queue.
func (c *VulkanContext) createLogicalDevice() error {
	queuePriority := []float32{1.0}

	queueCreateInfo := vk.DeviceQueueCreateInfo{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: c.computeFamily,
		QueueCount:       1,
		PQueuePriorities: queuePriority,
	}

	// No extensions needed for basic compute
	createInfo := vk.DeviceCreateInfo{
		SType:                vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount: 1,
		PQueueCreateInfos:    []vk.DeviceQueueCreateInfo{queueCreateInfo},
		// No device extensions required for compute
		EnabledExtensionCount:   0,
		PpEnabledExtensionNames: nil,
	}

	var device vk.Device
	if result := vk.CreateDevice(c.physicalDevice, &createInfo, nil, &device); result != vk.Success {
		return fmt.Errorf("failed to create logical device: %d", result)
	}
	c.device = device

	// Get compute queue
	var computeQueue vk.Queue
	vk.GetDeviceQueue(c.device, c.computeFamily, 0, &computeQueue)
	c.computeQueue = computeQueue

	// Cache memory properties for buffer allocation
	vk.GetPhysicalDeviceMemoryProperties(c.physicalDevice, &c.memoryProps)
	c.memoryProps.Deref()

	return nil
}

// createCommandPool creates a command pool for compute operations.
func (c *VulkanContext) createCommandPool() error {
	poolInfo := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: c.computeFamily,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
	}

	var commandPool vk.CommandPool
	if result := vk.CreateCommandPool(c.device, &poolInfo, nil, &commandPool); result != vk.Success {
		return fmt.Errorf("failed to create command pool: %d", result)
	}
	c.commandPool = commandPool

	return nil
}

// FindMemoryType finds a memory type that matches the filter and properties.
// Returns an error if no suitable memory type is found.
func (c *VulkanContext) FindMemoryType(typeFilter uint32, properties vk.MemoryPropertyFlags) (uint32, error) {
	for i := uint32(0); i < c.memoryProps.MemoryTypeCount; i++ {
		c.memoryProps.MemoryTypes[i].Deref()
		if (typeFilter&(1<<i)) != 0 &&
			(c.memoryProps.MemoryTypes[i].PropertyFlags&properties) == properties {
			return i, nil
		}
	}
	return 0, fmt.Errorf("failed to find suitable memory type")
}

// Destroy releases all Vulkan resources.
func (c *VulkanContext) Destroy() {
	if c.device != nil {
		vk.DeviceWaitIdle(c.device)

		if c.commandPool != nil {
			vk.DestroyCommandPool(c.device, c.commandPool, nil)
			c.commandPool = nil
		}

		vk.DestroyDevice(c.device, nil)
		c.device = nil
	}

	if c.instance != nil {
		vk.DestroyInstance(c.instance, nil)
		c.instance = nil
	}

	c.available = false
}

// Helper function for null-terminated strings.
func safeString(s string) string {
	if len(s) == 0 || s[len(s)-1] != 0 {
		return s + "\x00"
	}
	return s
}
