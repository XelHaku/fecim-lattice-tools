// Package render provides Vulkan-based visualization for ferroelectric simulations.
package render

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

// Vertex represents a vertex with position and color for GPU rendering.
type Vertex struct {
	Position [2]float32
	Color    [4]float32
}

// VulkanRenderer implements Vulkan-based rendering for the hysteresis visualizer.
type VulkanRenderer struct {
	config  *Config
	plot    *HysteresisPlot
	cell    *CellDisplay
	running bool

	// GLFW
	window *glfw.Window

	// Vulkan core
	instance       vk.Instance
	surface        vk.Surface
	physicalDevice vk.PhysicalDevice
	device         vk.Device
	graphicsQueue  vk.Queue
	presentQueue   vk.Queue

	// Swapchain
	swapchain       vk.Swapchain
	swapchainImages []vk.Image
	swapchainFormat vk.Format
	swapchainExtent vk.Extent2D
	imageViews      []vk.ImageView

	// Pipeline
	renderPass     vk.RenderPass
	pipelineLayout vk.PipelineLayout
	pipeline       vk.Pipeline
	linePipeline   vk.Pipeline // Pipeline for line rendering
	framebuffers   []vk.Framebuffer

	// Shaders
	vertShaderModule vk.ShaderModule
	fragShaderModule vk.ShaderModule

	// Vertex buffer
	vertexBuffer       vk.Buffer
	vertexBufferMemory vk.DeviceMemory
	vertexCount        uint32
	maxVertices        uint32

	// Commands
	commandPool    vk.CommandPool
	commandBuffers []vk.CommandBuffer

	// Sync
	imageAvailableSem vk.Semaphore
	renderFinishedSem vk.Semaphore
	inFlightFence     vk.Fence

	// Callbacks
	onUpdate func()

	// Queue family indices
	graphicsFamily uint32
	presentFamily  uint32

	// Memory type indices
	memoryProperties vk.PhysicalDeviceMemoryProperties

	// Shader path
	shaderDir string

	// Current level display (0-29)
	currentLevel int
}

// NewVulkanRenderer creates a new Vulkan-based renderer.
func NewVulkanRenderer(config *Config) *VulkanRenderer {
	// Find shader directory - try multiple locations
	shaderDir := "demo1-hysteresis/shaders"
	if _, err := os.Stat(shaderDir); os.IsNotExist(err) {
		// Try relative to executable
		exe, _ := os.Executable()
		exeDir := filepath.Dir(exe)
		shaderDir = filepath.Join(exeDir, "demo1-hysteresis", "shaders")
		if _, err := os.Stat(shaderDir); os.IsNotExist(err) {
			// Try runtime.Caller fallback
			_, filename, _, ok := runtime.Caller(0)
			if ok {
				pkgDir := filepath.Dir(filename)
				shaderDir = filepath.Join(pkgDir, "..", "..", "shaders")
			}
		}
	}

	return &VulkanRenderer{
		config:      config,
		cell:        NewCellDisplay(),
		maxVertices: 50000, // Support lots of curve points
		shaderDir:   shaderDir,
	}
}

// SetHysteresisPlot sets the plot to be rendered.
func (r *VulkanRenderer) SetHysteresisPlot(plot *HysteresisPlot) {
	r.plot = plot
}

// SetUpdateCallback sets a function to be called each frame.
func (r *VulkanRenderer) SetUpdateCallback(fn func()) {
	r.onUpdate = fn
}

// UpdatePolarization updates the cell polarization display.
func (r *VulkanRenderer) UpdatePolarization(normP float64) {
	r.cell.Polarization = normP
}

// Initialize sets up the Vulkan context and window.
func (r *VulkanRenderer) Initialize() error {
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		return fmt.Errorf("failed to initialize GLFW: %w", err)
	}

	// Initialize Vulkan
	procAddr := glfw.GetVulkanGetInstanceProcAddress()
	if procAddr == nil {
		return fmt.Errorf("vulkan not supported by GLFW")
	}
	vk.SetGetInstanceProcAddr(procAddr)

	if err := vk.Init(); err != nil {
		return fmt.Errorf("failed to initialize Vulkan: %w", err)
	}

	// Create window
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	glfw.WindowHint(glfw.Resizable, glfw.False)

	var err error
	r.window, err = glfw.CreateWindow(r.config.Width, r.config.Height, r.config.Title, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}

	// Create Vulkan instance
	if err := r.createInstance(); err != nil {
		return fmt.Errorf("failed to create Vulkan instance: %w", err)
	}

	// Create surface
	if err := r.createSurface(); err != nil {
		return fmt.Errorf("failed to create surface: %w", err)
	}

	// Pick physical device
	if err := r.pickPhysicalDevice(); err != nil {
		return fmt.Errorf("failed to pick physical device: %w", err)
	}

	// Create logical device
	if err := r.createLogicalDevice(); err != nil {
		return fmt.Errorf("failed to create logical device: %w", err)
	}

	// Create swapchain
	if err := r.createSwapchain(); err != nil {
		return fmt.Errorf("failed to create swapchain: %w", err)
	}

	// Create image views
	if err := r.createImageViews(); err != nil {
		return fmt.Errorf("failed to create image views: %w", err)
	}

	// Create render pass
	if err := r.createRenderPass(); err != nil {
		return fmt.Errorf("failed to create render pass: %w", err)
	}

	// Create pipeline
	if err := r.createGraphicsPipeline(); err != nil {
		return fmt.Errorf("failed to create graphics pipeline: %w", err)
	}

	// Create framebuffers
	if err := r.createFramebuffers(); err != nil {
		return fmt.Errorf("failed to create framebuffers: %w", err)
	}

	// Create command pool
	if err := r.createCommandPool(); err != nil {
		return fmt.Errorf("failed to create command pool: %w", err)
	}

	// Create command buffers
	if err := r.createCommandBuffers(); err != nil {
		return fmt.Errorf("failed to create command buffers: %w", err)
	}

	// Create sync objects
	if err := r.createSyncObjects(); err != nil {
		return fmt.Errorf("failed to create sync objects: %w", err)
	}

	// Create vertex buffer
	if err := r.createVertexBuffer(); err != nil {
		return fmt.Errorf("failed to create vertex buffer: %w", err)
	}

	return nil
}

func (r *VulkanRenderer) createInstance() error {
	// Get required extensions from GLFW
	requiredExts := r.window.GetRequiredInstanceExtensions()

	appInfo := vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PApplicationName:   safeString("IronLattice Hysteresis"),
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        safeString("IronLattice"),
		EngineVersion:      vk.MakeVersion(1, 0, 0),
		ApiVersion:         vk.ApiVersion11,
	}

	createInfo := vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo:        &appInfo,
		EnabledExtensionCount:   uint32(len(requiredExts)),
		PpEnabledExtensionNames: requiredExts,
	}

	var instance vk.Instance
	if result := vk.CreateInstance(&createInfo, nil, &instance); result != vk.Success {
		return fmt.Errorf("vkCreateInstance failed: %d", result)
	}
	r.instance = instance

	vk.InitInstance(r.instance)
	return nil
}

func (r *VulkanRenderer) createSurface() error {
	surfacePtr, err := r.window.CreateWindowSurface(r.instance, nil)
	if err != nil {
		return err
	}
	r.surface = vk.SurfaceFromPointer(surfacePtr)
	return nil
}

func (r *VulkanRenderer) pickPhysicalDevice() error {
	var deviceCount uint32
	vk.EnumeratePhysicalDevices(r.instance, &deviceCount, nil)
	if deviceCount == 0 {
		return fmt.Errorf("no Vulkan-capable GPU found")
	}

	devices := make([]vk.PhysicalDevice, deviceCount)
	vk.EnumeratePhysicalDevices(r.instance, &deviceCount, devices)

	for _, device := range devices {
		if r.isDeviceSuitable(device) {
			r.physicalDevice = device
			return nil
		}
	}

	return fmt.Errorf("no suitable GPU found")
}

func (r *VulkanRenderer) isDeviceSuitable(device vk.PhysicalDevice) bool {
	// Check queue families
	graphicsFamily, presentFamily, found := r.findQueueFamilies(device)
	if !found {
		return false
	}
	r.graphicsFamily = graphicsFamily
	r.presentFamily = presentFamily

	// Check swapchain support
	var formatCount, presentModeCount uint32
	vk.GetPhysicalDeviceSurfaceFormats(device, r.surface, &formatCount, nil)
	vk.GetPhysicalDeviceSurfacePresentModes(device, r.surface, &presentModeCount, nil)

	return formatCount > 0 && presentModeCount > 0
}

func (r *VulkanRenderer) findQueueFamilies(device vk.PhysicalDevice) (graphics, present uint32, found bool) {
	var queueFamilyCount uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, nil)
	queueFamilies := make([]vk.QueueFamilyProperties, queueFamilyCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, queueFamilies)

	graphicsFound, presentFound := false, false

	for i, qf := range queueFamilies {
		qf.Deref()
		if qf.QueueFlags&vk.QueueFlags(vk.QueueGraphicsBit) != 0 {
			graphics = uint32(i)
			graphicsFound = true
		}

		var presentSupport vk.Bool32
		vk.GetPhysicalDeviceSurfaceSupport(device, uint32(i), r.surface, &presentSupport)
		if presentSupport == vk.True {
			present = uint32(i)
			presentFound = true
		}

		if graphicsFound && presentFound {
			return graphics, present, true
		}
	}

	return 0, 0, false
}

func (r *VulkanRenderer) createLogicalDevice() error {
	uniqueFamilies := make(map[uint32]bool)
	uniqueFamilies[r.graphicsFamily] = true
	uniqueFamilies[r.presentFamily] = true

	var queueCreateInfos []vk.DeviceQueueCreateInfo
	queuePriority := []float32{1.0}

	for family := range uniqueFamilies {
		queueCreateInfo := vk.DeviceQueueCreateInfo{
			SType:            vk.StructureTypeDeviceQueueCreateInfo,
			QueueFamilyIndex: family,
			QueueCount:       1,
			PQueuePriorities: queuePriority,
		}
		queueCreateInfos = append(queueCreateInfos, queueCreateInfo)
	}

	deviceExtensions := []string{
		vk.KhrSwapchainExtensionName + "\x00",
	}

	createInfo := vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(queueCreateInfos)),
		PQueueCreateInfos:       queueCreateInfos,
		EnabledExtensionCount:   uint32(len(deviceExtensions)),
		PpEnabledExtensionNames: deviceExtensions,
	}

	var device vk.Device
	if result := vk.CreateDevice(r.physicalDevice, &createInfo, nil, &device); result != vk.Success {
		return fmt.Errorf("failed to create logical device: %d", result)
	}
	r.device = device

	var graphicsQueue, presentQueue vk.Queue
	vk.GetDeviceQueue(r.device, r.graphicsFamily, 0, &graphicsQueue)
	vk.GetDeviceQueue(r.device, r.presentFamily, 0, &presentQueue)
	r.graphicsQueue = graphicsQueue
	r.presentQueue = presentQueue

	// Get memory properties for buffer allocation
	vk.GetPhysicalDeviceMemoryProperties(r.physicalDevice, &r.memoryProperties)
	r.memoryProperties.Deref()

	return nil
}

func (r *VulkanRenderer) createSwapchain() error {
	var capabilities vk.SurfaceCapabilities
	vk.GetPhysicalDeviceSurfaceCapabilities(r.physicalDevice, r.surface, &capabilities)
	capabilities.Deref()

	// Choose format
	var formatCount uint32
	vk.GetPhysicalDeviceSurfaceFormats(r.physicalDevice, r.surface, &formatCount, nil)
	formats := make([]vk.SurfaceFormat, formatCount)
	vk.GetPhysicalDeviceSurfaceFormats(r.physicalDevice, r.surface, &formatCount, formats)

	chosenFormat := formats[0]
	for _, format := range formats {
		format.Deref()
		if format.Format == vk.FormatB8g8r8a8Srgb && format.ColorSpace == vk.ColorSpaceSrgbNonlinear {
			chosenFormat = format
			break
		}
	}
	chosenFormat.Deref()
	r.swapchainFormat = chosenFormat.Format

	// Choose extent
	r.swapchainExtent = capabilities.CurrentExtent
	r.swapchainExtent.Deref()

	imageCount := capabilities.MinImageCount + 1
	if capabilities.MaxImageCount > 0 && imageCount > capabilities.MaxImageCount {
		imageCount = capabilities.MaxImageCount
	}

	createInfo := vk.SwapchainCreateInfo{
		SType:            vk.StructureTypeSwapchainCreateInfo,
		Surface:          r.surface,
		MinImageCount:    imageCount,
		ImageFormat:      chosenFormat.Format,
		ImageColorSpace:  chosenFormat.ColorSpace,
		ImageExtent:      r.swapchainExtent,
		ImageArrayLayers: 1,
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		PreTransform:     capabilities.CurrentTransform,
		CompositeAlpha:   vk.CompositeAlphaOpaqueBit,
		PresentMode:      vk.PresentModeFifo,
		Clipped:          vk.True,
	}

	if r.graphicsFamily != r.presentFamily {
		createInfo.ImageSharingMode = vk.SharingModeConcurrent
		createInfo.QueueFamilyIndexCount = 2
		createInfo.PQueueFamilyIndices = []uint32{r.graphicsFamily, r.presentFamily}
	} else {
		createInfo.ImageSharingMode = vk.SharingModeExclusive
	}

	var swapchain vk.Swapchain
	if result := vk.CreateSwapchain(r.device, &createInfo, nil, &swapchain); result != vk.Success {
		return fmt.Errorf("failed to create swapchain: %d", result)
	}
	r.swapchain = swapchain

	// Get swapchain images
	vk.GetSwapchainImages(r.device, r.swapchain, &imageCount, nil)
	r.swapchainImages = make([]vk.Image, imageCount)
	vk.GetSwapchainImages(r.device, r.swapchain, &imageCount, r.swapchainImages)

	return nil
}

func (r *VulkanRenderer) createImageViews() error {
	r.imageViews = make([]vk.ImageView, len(r.swapchainImages))

	for i, image := range r.swapchainImages {
		createInfo := vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			Image:    image,
			ViewType: vk.ImageViewType2d,
			Format:   r.swapchainFormat,
			Components: vk.ComponentMapping{
				R: vk.ComponentSwizzleIdentity,
				G: vk.ComponentSwizzleIdentity,
				B: vk.ComponentSwizzleIdentity,
				A: vk.ComponentSwizzleIdentity,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		}

		var imageView vk.ImageView
		if result := vk.CreateImageView(r.device, &createInfo, nil, &imageView); result != vk.Success {
			return fmt.Errorf("failed to create image view: %d", result)
		}
		r.imageViews[i] = imageView
	}

	return nil
}

func (r *VulkanRenderer) createRenderPass() error {
	colorAttachment := vk.AttachmentDescription{
		Format:         r.swapchainFormat,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutPresentSrc,
	}

	colorAttachmentRef := vk.AttachmentReference{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}

	subpass := vk.SubpassDescription{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: 1,
		PColorAttachments:    []vk.AttachmentReference{colorAttachmentRef},
	}

	dependency := vk.SubpassDependency{
		SrcSubpass:    vk.SubpassExternal,
		DstSubpass:    0,
		SrcStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		SrcAccessMask: 0,
		DstStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		DstAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
	}

	renderPassInfo := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    []vk.AttachmentDescription{colorAttachment},
		SubpassCount:    1,
		PSubpasses:      []vk.SubpassDescription{subpass},
		DependencyCount: 1,
		PDependencies:   []vk.SubpassDependency{dependency},
	}

	var renderPass vk.RenderPass
	if result := vk.CreateRenderPass(r.device, &renderPassInfo, nil, &renderPass); result != vk.Success {
		return fmt.Errorf("failed to create render pass: %d", result)
	}
	r.renderPass = renderPass

	return nil
}

func (r *VulkanRenderer) createGraphicsPipeline() error {
	// Load shader modules
	vertPath := filepath.Join(r.shaderDir, "simple.vert.spv")
	fragPath := filepath.Join(r.shaderDir, "simple.frag.spv")

	vertCode, err := os.ReadFile(vertPath)
	if err != nil {
		return fmt.Errorf("failed to read vertex shader: %w", err)
	}

	fragCode, err := os.ReadFile(fragPath)
	if err != nil {
		return fmt.Errorf("failed to read fragment shader: %w", err)
	}

	// Create shader modules
	r.vertShaderModule, err = r.createShaderModule(vertCode)
	if err != nil {
		return fmt.Errorf("failed to create vertex shader module: %w", err)
	}

	r.fragShaderModule, err = r.createShaderModule(fragCode)
	if err != nil {
		return fmt.Errorf("failed to create fragment shader module: %w", err)
	}

	// Shader stage info
	vertShaderStageInfo := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageVertexBit,
		Module: r.vertShaderModule,
		PName:  safeString("main"),
	}

	fragShaderStageInfo := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageFragmentBit,
		Module: r.fragShaderModule,
		PName:  safeString("main"),
	}

	shaderStages := []vk.PipelineShaderStageCreateInfo{vertShaderStageInfo, fragShaderStageInfo}

	// Vertex input - position (vec2) + color (vec4)
	bindingDescription := vk.VertexInputBindingDescription{
		Binding:   0,
		Stride:    uint32(unsafe.Sizeof(Vertex{})),
		InputRate: vk.VertexInputRateVertex,
	}

	attributeDescriptions := []vk.VertexInputAttributeDescription{
		{
			Binding:  0,
			Location: 0,
			Format:   vk.FormatR32g32Sfloat, // vec2 position
			Offset:   0,
		},
		{
			Binding:  0,
			Location: 1,
			Format:   vk.FormatR32g32b32a32Sfloat, // vec4 color
			Offset:   uint32(unsafe.Offsetof(Vertex{}.Color)),
		},
	}

	vertexInputInfo := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   1,
		PVertexBindingDescriptions:      []vk.VertexInputBindingDescription{bindingDescription},
		VertexAttributeDescriptionCount: uint32(len(attributeDescriptions)),
		PVertexAttributeDescriptions:    attributeDescriptions,
	}

	// Input assembly - triangles for fills, lines for curves
	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               vk.PrimitiveTopologyTriangleList,
		PrimitiveRestartEnable: vk.False,
	}

	// Viewport and scissor
	viewport := vk.Viewport{
		X:        0.0,
		Y:        0.0,
		Width:    float32(r.swapchainExtent.Width),
		Height:   float32(r.swapchainExtent.Height),
		MinDepth: 0.0,
		MaxDepth: 1.0,
	}

	scissor := vk.Rect2D{
		Offset: vk.Offset2D{X: 0, Y: 0},
		Extent: r.swapchainExtent,
	}

	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		PViewports:    []vk.Viewport{viewport},
		ScissorCount:  1,
		PScissors:     []vk.Rect2D{scissor},
	}

	// Rasterizer
	rasterizer := vk.PipelineRasterizationStateCreateInfo{
		SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
		DepthClampEnable:        vk.False,
		RasterizerDiscardEnable: vk.False,
		PolygonMode:             vk.PolygonModeFill,
		LineWidth:               1.0,
		CullMode:                vk.CullModeFlags(vk.CullModeNone),
		FrontFace:               vk.FrontFaceCounterClockwise,
		DepthBiasEnable:         vk.False,
	}

	// Multisampling
	multisampling := vk.PipelineMultisampleStateCreateInfo{
		SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
		SampleShadingEnable:  vk.False,
		RasterizationSamples: vk.SampleCount1Bit,
	}

	// Color blending with alpha support
	colorBlendAttachment := vk.PipelineColorBlendAttachmentState{
		ColorWriteMask:      vk.ColorComponentFlags(vk.ColorComponentRBit | vk.ColorComponentGBit | vk.ColorComponentBBit | vk.ColorComponentABit),
		BlendEnable:         vk.True,
		SrcColorBlendFactor: vk.BlendFactorSrcAlpha,
		DstColorBlendFactor: vk.BlendFactorOneMinusSrcAlpha,
		ColorBlendOp:        vk.BlendOpAdd,
		SrcAlphaBlendFactor: vk.BlendFactorOne,
		DstAlphaBlendFactor: vk.BlendFactorZero,
		AlphaBlendOp:        vk.BlendOpAdd,
	}

	colorBlending := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   vk.False,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{colorBlendAttachment},
	}

	// Pipeline layout
	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{
		SType: vk.StructureTypePipelineLayoutCreateInfo,
	}

	var pipelineLayout vk.PipelineLayout
	if result := vk.CreatePipelineLayout(r.device, &pipelineLayoutInfo, nil, &pipelineLayout); result != vk.Success {
		return fmt.Errorf("failed to create pipeline layout: %d", result)
	}
	r.pipelineLayout = pipelineLayout

	// Create the graphics pipeline for triangles
	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          uint32(len(shaderStages)),
		PStages:             shaderStages,
		PVertexInputState:   &vertexInputInfo,
		PInputAssemblyState: &inputAssembly,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterizer,
		PMultisampleState:   &multisampling,
		PColorBlendState:    &colorBlending,
		Layout:              r.pipelineLayout,
		RenderPass:          r.renderPass,
		Subpass:             0,
	}

	pipelines := make([]vk.Pipeline, 1)
	if result := vk.CreateGraphicsPipelines(r.device, vk.NullPipelineCache, 1, []vk.GraphicsPipelineCreateInfo{pipelineInfo}, nil, pipelines); result != vk.Success {
		return fmt.Errorf("failed to create graphics pipeline: %d", result)
	}
	r.pipeline = pipelines[0]

	// Create line pipeline for curves
	lineAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               vk.PrimitiveTopologyLineList,
		PrimitiveRestartEnable: vk.False,
	}

	lineRasterizer := vk.PipelineRasterizationStateCreateInfo{
		SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
		DepthClampEnable:        vk.False,
		RasterizerDiscardEnable: vk.False,
		PolygonMode:             vk.PolygonModeFill,
		LineWidth:               2.0, // Thicker lines for curves
		CullMode:                vk.CullModeFlags(vk.CullModeNone),
		FrontFace:               vk.FrontFaceCounterClockwise,
		DepthBiasEnable:         vk.False,
	}

	linePipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          uint32(len(shaderStages)),
		PStages:             shaderStages,
		PVertexInputState:   &vertexInputInfo,
		PInputAssemblyState: &lineAssembly,
		PViewportState:      &viewportState,
		PRasterizationState: &lineRasterizer,
		PMultisampleState:   &multisampling,
		PColorBlendState:    &colorBlending,
		Layout:              r.pipelineLayout,
		RenderPass:          r.renderPass,
		Subpass:             0,
	}

	linePipelines := make([]vk.Pipeline, 1)
	if result := vk.CreateGraphicsPipelines(r.device, vk.NullPipelineCache, 1, []vk.GraphicsPipelineCreateInfo{linePipelineInfo}, nil, linePipelines); result != vk.Success {
		return fmt.Errorf("failed to create line graphics pipeline: %d", result)
	}
	r.linePipeline = linePipelines[0]

	return nil
}

func (r *VulkanRenderer) createShaderModule(code []byte) (vk.ShaderModule, error) {
	// Convert to uint32 slice
	codeUint32 := make([]uint32, len(code)/4)
	for i := range codeUint32 {
		codeUint32[i] = binary.LittleEndian.Uint32(code[i*4:])
	}

	createInfo := vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len(code)),
		PCode:    codeUint32,
	}

	var shaderModule vk.ShaderModule
	if result := vk.CreateShaderModule(r.device, &createInfo, nil, &shaderModule); result != vk.Success {
		return nil, fmt.Errorf("failed to create shader module: %d", result)
	}

	return shaderModule, nil
}

func (r *VulkanRenderer) createFramebuffers() error {
	r.framebuffers = make([]vk.Framebuffer, len(r.imageViews))

	for i, imageView := range r.imageViews {
		attachments := []vk.ImageView{imageView}

		framebufferInfo := vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      r.renderPass,
			AttachmentCount: 1,
			PAttachments:    attachments,
			Width:           r.swapchainExtent.Width,
			Height:          r.swapchainExtent.Height,
			Layers:          1,
		}

		var framebuffer vk.Framebuffer
		if result := vk.CreateFramebuffer(r.device, &framebufferInfo, nil, &framebuffer); result != vk.Success {
			return fmt.Errorf("failed to create framebuffer: %d", result)
		}
		r.framebuffers[i] = framebuffer
	}

	return nil
}

func (r *VulkanRenderer) createCommandPool() error {
	poolInfo := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: r.graphicsFamily,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
	}

	var commandPool vk.CommandPool
	if result := vk.CreateCommandPool(r.device, &poolInfo, nil, &commandPool); result != vk.Success {
		return fmt.Errorf("failed to create command pool: %d", result)
	}
	r.commandPool = commandPool

	return nil
}

func (r *VulkanRenderer) createCommandBuffers() error {
	r.commandBuffers = make([]vk.CommandBuffer, len(r.framebuffers))

	allocInfo := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        r.commandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: uint32(len(r.commandBuffers)),
	}

	if result := vk.AllocateCommandBuffers(r.device, &allocInfo, r.commandBuffers); result != vk.Success {
		return fmt.Errorf("failed to allocate command buffers: %d", result)
	}

	return nil
}

func (r *VulkanRenderer) createSyncObjects() error {
	semaphoreInfo := vk.SemaphoreCreateInfo{
		SType: vk.StructureTypeSemaphoreCreateInfo,
	}

	fenceInfo := vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vk.FenceCreateSignaledBit),
	}

	var imageAvailableSem, renderFinishedSem vk.Semaphore
	var inFlightFence vk.Fence

	if vk.CreateSemaphore(r.device, &semaphoreInfo, nil, &imageAvailableSem) != vk.Success ||
		vk.CreateSemaphore(r.device, &semaphoreInfo, nil, &renderFinishedSem) != vk.Success ||
		vk.CreateFence(r.device, &fenceInfo, nil, &inFlightFence) != vk.Success {
		return fmt.Errorf("failed to create sync objects")
	}

	r.imageAvailableSem = imageAvailableSem
	r.renderFinishedSem = renderFinishedSem
	r.inFlightFence = inFlightFence

	return nil
}

func (r *VulkanRenderer) createVertexBuffer() error {
	bufferSize := vk.DeviceSize(r.maxVertices * uint32(unsafe.Sizeof(Vertex{})))

	bufferInfo := vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        bufferSize,
		Usage:       vk.BufferUsageFlags(vk.BufferUsageVertexBufferBit),
		SharingMode: vk.SharingModeExclusive,
	}

	var buffer vk.Buffer
	if result := vk.CreateBuffer(r.device, &bufferInfo, nil, &buffer); result != vk.Success {
		return fmt.Errorf("failed to create vertex buffer: %d", result)
	}
	r.vertexBuffer = buffer

	// Get memory requirements
	var memRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(r.device, r.vertexBuffer, &memRequirements)
	memRequirements.Deref()

	// Find suitable memory type
	memTypeIndex := r.findMemoryType(memRequirements.MemoryTypeBits,
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if memTypeIndex == ^uint32(0) {
		return fmt.Errorf("failed to find suitable memory type")
	}

	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memRequirements.Size,
		MemoryTypeIndex: memTypeIndex,
	}

	var memory vk.DeviceMemory
	if result := vk.AllocateMemory(r.device, &allocInfo, nil, &memory); result != vk.Success {
		return fmt.Errorf("failed to allocate vertex buffer memory: %d", result)
	}
	r.vertexBufferMemory = memory

	vk.BindBufferMemory(r.device, r.vertexBuffer, r.vertexBufferMemory, 0)

	return nil
}

func (r *VulkanRenderer) findMemoryType(typeFilter uint32, properties vk.MemoryPropertyFlags) uint32 {
	for i := uint32(0); i < r.memoryProperties.MemoryTypeCount; i++ {
		r.memoryProperties.MemoryTypes[i].Deref()
		if (typeFilter&(1<<i)) != 0 &&
			(r.memoryProperties.MemoryTypes[i].PropertyFlags&properties) == properties {
			return i
		}
	}
	return ^uint32(0)
}

func (r *VulkanRenderer) updateVertexBuffer(vertices []Vertex) error {
	if len(vertices) == 0 {
		r.vertexCount = 0
		return nil
	}

	if uint32(len(vertices)) > r.maxVertices {
		vertices = vertices[:r.maxVertices]
	}

	bufferSize := vk.DeviceSize(len(vertices) * int(unsafe.Sizeof(Vertex{})))

	var data unsafe.Pointer
	vk.MapMemory(r.device, r.vertexBufferMemory, 0, bufferSize, 0, &data)

	// Copy vertex data
	vertexSlice := (*[1 << 30]Vertex)(data)[:len(vertices):len(vertices)]
	copy(vertexSlice, vertices)

	vk.UnmapMemory(r.device, r.vertexBufferMemory)

	r.vertexCount = uint32(len(vertices))
	return nil
}

func (r *VulkanRenderer) recordCommandBuffer(commandBuffer vk.CommandBuffer, imageIndex uint32) error {
	beginInfo := vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
	}

	if result := vk.BeginCommandBuffer(commandBuffer, &beginInfo); result != vk.Success {
		return fmt.Errorf("failed to begin command buffer: %d", result)
	}

	// Dark background
	clearValue := vk.ClearValue{}
	clearValue.SetColor([]float32{0.05, 0.05, 0.08, 1.0})

	renderPassInfo := vk.RenderPassBeginInfo{
		SType:       vk.StructureTypeRenderPassBeginInfo,
		RenderPass:  r.renderPass,
		Framebuffer: r.framebuffers[imageIndex],
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{X: 0, Y: 0},
			Extent: r.swapchainExtent,
		},
		ClearValueCount: 1,
		PClearValues:    []vk.ClearValue{clearValue},
	}

	vk.CmdBeginRenderPass(commandBuffer, &renderPassInfo, vk.SubpassContentsInline)

	// Generate vertices for visualization
	vertices := r.generateVisualizationVertices()
	if len(vertices) > 0 {
		r.updateVertexBuffer(vertices)

		// Bind pipeline and draw
		vertexBuffers := []vk.Buffer{r.vertexBuffer}
		offsets := []vk.DeviceSize{0}

		// Draw filled primitives (cell, background elements)
		vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, r.pipeline)
		vk.CmdBindVertexBuffers(commandBuffer, 0, 1, vertexBuffers, offsets)

		// Count triangle vertices (groups of 3)
		triangleCount := r.countTriangleVertices(vertices)
		if triangleCount > 0 {
			vk.CmdDraw(commandBuffer, triangleCount, 1, 0, 0)
		}

		// Draw lines (curves)
		if r.linePipeline != nil && r.vertexCount > triangleCount {
			vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, r.linePipeline)
			lineCount := r.vertexCount - triangleCount
			vk.CmdDraw(commandBuffer, lineCount, 1, triangleCount, 0)
		}
	}

	vk.CmdEndRenderPass(commandBuffer)

	if result := vk.EndCommandBuffer(commandBuffer); result != vk.Success {
		return fmt.Errorf("failed to end command buffer: %d", result)
	}

	return nil
}

func (r *VulkanRenderer) countTriangleVertices(vertices []Vertex) uint32 {
	// Count vertices that are part of filled triangles
	// Cell: 6, Plot bg: 6, Level bar bg: 6, Level bar segments: 30*6 = 180
	// Total triangle vertices: 6 + 6 + 6 + 180 = 198
	return 198
}

func (r *VulkanRenderer) generateVisualizationVertices() []Vertex {
	var vertices []Vertex

	// Get current polarization and compute level (0-29)
	normP := r.cell.Polarization
	level := int(math.Round((normP + 1.0) / 2.0 * 29.0))
	if level < 0 {
		level = 0
	}
	if level > 29 {
		level = 29
	}
	r.currentLevel = level

	// === TRIANGLES FIRST (drawn with triangle pipeline) ===

	// 1. Draw ferroelectric cell (colored rectangle on left side) - 6 vertices
	cellVertices := r.generateCellVertices(normP, level)
	vertices = append(vertices, cellVertices...)

	// 2. Draw P-E plot area background - 6 vertices
	plotBgVertices := r.generatePlotBackgroundVertices()
	vertices = append(vertices, plotBgVertices...)

	// 3. Draw level indicator bar - 6 + 180 = 186 vertices
	levelVertices := r.generateLevelBarVertices(level)
	vertices = append(vertices, levelVertices...)

	// === LINES AFTER (drawn with line pipeline) ===

	// 4. Draw grid lines
	gridVertices := r.generateGridVertices()
	vertices = append(vertices, gridVertices...)

	// 5. Draw axes
	axisVertices := r.generateAxisVertices()
	vertices = append(vertices, axisVertices...)

	// 6. Draw hysteresis curve from plot data
	if r.plot != nil && len(r.plot.Points) > 1 {
		curveVertices := r.generateCurveVertices()
		vertices = append(vertices, curveVertices...)
	}

	// 7. Draw current position marker
	if r.plot != nil {
		markerVertices := r.generateMarkerVertices()
		vertices = append(vertices, markerVertices...)
	}

	return vertices
}

func (r *VulkanRenderer) generateCellVertices(normP float64, level int) []Vertex {
	// Cell position (left side of screen)
	x1 := float32(-0.9)
	y1 := float32(-0.5)
	x2 := float32(-0.5)
	y2 := float32(0.5)

	// Color based on polarization: blue (-) to white (0) to red (+)
	var color [4]float32
	if normP >= 0 {
		// White to red
		t := float32(normP)
		color = [4]float32{1.0, 1.0 - t*0.7, 1.0 - t*0.9, 1.0}
	} else {
		// White to blue
		t := float32(-normP)
		color = [4]float32{1.0 - t*0.9, 1.0 - t*0.7, 1.0, 1.0}
	}

	// Two triangles forming a quad
	return []Vertex{
		{Position: [2]float32{x1, y1}, Color: color},
		{Position: [2]float32{x2, y1}, Color: color},
		{Position: [2]float32{x1, y2}, Color: color},
		{Position: [2]float32{x2, y1}, Color: color},
		{Position: [2]float32{x2, y2}, Color: color},
		{Position: [2]float32{x1, y2}, Color: color},
	}
}

func (r *VulkanRenderer) generatePlotBackgroundVertices() []Vertex {
	// Plot area background (dark)
	x1 := float32(-0.35)
	y1 := float32(-0.8)
	x2 := float32(0.95)
	y2 := float32(0.8)

	bgColor := [4]float32{0.1, 0.1, 0.12, 1.0}

	return []Vertex{
		{Position: [2]float32{x1, y1}, Color: bgColor},
		{Position: [2]float32{x2, y1}, Color: bgColor},
		{Position: [2]float32{x1, y2}, Color: bgColor},
		{Position: [2]float32{x2, y1}, Color: bgColor},
		{Position: [2]float32{x2, y2}, Color: bgColor},
		{Position: [2]float32{x1, y2}, Color: bgColor},
	}
}

func (r *VulkanRenderer) generateGridVertices() []Vertex {
	var vertices []Vertex

	gridColor := [4]float32{0.25, 0.25, 0.3, 0.5}

	// Plot bounds
	plotX1, plotY1 := float32(-0.35), float32(-0.8)
	plotX2, plotY2 := float32(0.95), float32(0.8)

	// Grid lines (10 divisions)
	divisions := 10
	for i := 0; i <= divisions; i++ {
		t := float32(i) / float32(divisions)

		// Vertical line
		x := plotX1 + t*(plotX2-plotX1)
		vertices = append(vertices,
			Vertex{Position: [2]float32{x, plotY1}, Color: gridColor},
			Vertex{Position: [2]float32{x, plotY2}, Color: gridColor},
		)

		// Horizontal line
		y := plotY1 + t*(plotY2-plotY1)
		vertices = append(vertices,
			Vertex{Position: [2]float32{plotX1, y}, Color: gridColor},
			Vertex{Position: [2]float32{plotX2, y}, Color: gridColor},
		)
	}

	return vertices
}

func (r *VulkanRenderer) generateAxisVertices() []Vertex {
	axisColor := [4]float32{0.8, 0.8, 0.8, 1.0}

	// Plot bounds
	plotX1, plotY1 := float32(-0.35), float32(-0.8)
	plotX2, plotY2 := float32(0.95), float32(0.8)

	// Center of plot
	centerX := (plotX1 + plotX2) / 2
	centerY := (plotY1 + plotY2) / 2

	return []Vertex{
		// X-axis (E field)
		{Position: [2]float32{plotX1, centerY}, Color: axisColor},
		{Position: [2]float32{plotX2, centerY}, Color: axisColor},
		// Y-axis (Polarization)
		{Position: [2]float32{centerX, plotY1}, Color: axisColor},
		{Position: [2]float32{centerX, plotY2}, Color: axisColor},
	}
}

func (r *VulkanRenderer) generateCurveVertices() []Vertex {
	var vertices []Vertex

	if r.plot == nil || len(r.plot.Points) < 2 {
		return vertices
	}

	curveColor := [4]float32{0.2, 0.6, 1.0, 1.0}

	// Plot bounds
	plotX1, plotY1 := float32(-0.35), float32(-0.8)
	plotX2, plotY2 := float32(0.95), float32(0.8)

	// Convert plot points to screen coordinates
	for i := 0; i < len(r.plot.Points)-1; i++ {
		// Normalize to 0-1
		x1, y1 := r.plot.NormalizeToScreen(r.plot.Points[i].X, r.plot.Points[i].Y)
		x2, y2 := r.plot.NormalizeToScreen(r.plot.Points[i+1].X, r.plot.Points[i+1].Y)

		// Convert to plot area coordinates
		screenX1 := plotX1 + float32(x1)*(plotX2-plotX1)
		screenY1 := plotY1 + float32(y1)*(plotY2-plotY1)
		screenX2 := plotX1 + float32(x2)*(plotX2-plotX1)
		screenY2 := plotY1 + float32(y2)*(plotY2-plotY1)

		// Trail fade effect - older points are more transparent
		alpha := float32(0.3) + float32(i)/float32(len(r.plot.Points))*0.7
		fadeColor := [4]float32{curveColor[0], curveColor[1], curveColor[2], alpha}

		vertices = append(vertices,
			Vertex{Position: [2]float32{screenX1, screenY1}, Color: fadeColor},
			Vertex{Position: [2]float32{screenX2, screenY2}, Color: fadeColor},
		)
	}

	return vertices
}

func (r *VulkanRenderer) generateMarkerVertices() []Vertex {
	var vertices []Vertex

	if r.plot == nil {
		return vertices
	}

	markerColor := [4]float32{1.0, 0.3, 0.3, 1.0}

	// Plot bounds
	plotX1, plotY1 := float32(-0.35), float32(-0.8)
	plotX2, plotY2 := float32(0.95), float32(0.8)

	// Current position
	x, y := r.plot.NormalizeToScreen(r.plot.CurrentE, r.plot.CurrentP)
	screenX := plotX1 + float32(x)*(plotX2-plotX1)
	screenY := plotY1 + float32(y)*(plotY2-plotY1)

	// Draw a diamond marker
	size := float32(0.02)
	vertices = append(vertices,
		// Top
		Vertex{Position: [2]float32{screenX, screenY + size}, Color: markerColor},
		Vertex{Position: [2]float32{screenX - size, screenY}, Color: markerColor},
		// Right
		Vertex{Position: [2]float32{screenX, screenY + size}, Color: markerColor},
		Vertex{Position: [2]float32{screenX + size, screenY}, Color: markerColor},
		// Bottom
		Vertex{Position: [2]float32{screenX + size, screenY}, Color: markerColor},
		Vertex{Position: [2]float32{screenX, screenY - size}, Color: markerColor},
		// Left
		Vertex{Position: [2]float32{screenX, screenY - size}, Color: markerColor},
		Vertex{Position: [2]float32{screenX - size, screenY}, Color: markerColor},
	)

	return vertices
}

func (r *VulkanRenderer) generateLevelBarVertices(level int) []Vertex {
	var vertices []Vertex

	// Level indicator bar on the right of the cell
	barX := float32(-0.45)
	barY1 := float32(-0.5)
	barY2 := float32(0.5)
	barWidth := float32(0.03)

	// Background bar
	bgColor := [4]float32{0.2, 0.2, 0.25, 1.0}
	vertices = append(vertices,
		Vertex{Position: [2]float32{barX, barY1}, Color: bgColor},
		Vertex{Position: [2]float32{barX + barWidth, barY1}, Color: bgColor},
		Vertex{Position: [2]float32{barX, barY2}, Color: bgColor},
		Vertex{Position: [2]float32{barX + barWidth, barY1}, Color: bgColor},
		Vertex{Position: [2]float32{barX + barWidth, barY2}, Color: bgColor},
		Vertex{Position: [2]float32{barX, barY2}, Color: bgColor},
	)

	// Level indicator segments (30 levels)
	segHeight := (barY2 - barY1) / 30.0
	for i := 0; i < 30; i++ {
		y1 := barY1 + float32(i)*segHeight + 0.002
		y2 := barY1 + float32(i+1)*segHeight - 0.002

		var segColor [4]float32
		if i == level {
			// Current level - bright white
			segColor = [4]float32{1.0, 1.0, 1.0, 1.0}
		} else if i < level {
			// Below current level - gradient blue to white
			t := float32(i) / 29.0
			segColor = [4]float32{0.3 + t*0.7, 0.3 + t*0.7, 1.0, 0.6}
		} else {
			// Above current level - gradient white to red
			t := float32(i) / 29.0
			segColor = [4]float32{1.0, 1.0 - t*0.7, 1.0 - t*0.7, 0.6}
		}

		vertices = append(vertices,
			Vertex{Position: [2]float32{barX + 0.002, y1}, Color: segColor},
			Vertex{Position: [2]float32{barX + barWidth - 0.002, y1}, Color: segColor},
			Vertex{Position: [2]float32{barX + 0.002, y2}, Color: segColor},
			Vertex{Position: [2]float32{barX + barWidth - 0.002, y1}, Color: segColor},
			Vertex{Position: [2]float32{barX + barWidth - 0.002, y2}, Color: segColor},
			Vertex{Position: [2]float32{barX + 0.002, y2}, Color: segColor},
		)
	}

	return vertices
}

// Run starts the main render loop.
func (r *VulkanRenderer) Run() error {
	r.running = true

	// Set up key callback for interactive control
	r.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press || action == glfw.Repeat {
			switch key {
			case glfw.KeyEscape:
				r.running = false
			case glfw.KeyQ:
				r.running = false
			}
		}
	})

	// Print controls
	fmt.Println("\nControls:")
	fmt.Println("  ESC/Q : Quit")
	fmt.Println("  The simulation runs automatically")
	fmt.Println()

	for r.running && !r.window.ShouldClose() {
		glfw.PollEvents()

		// Call update callback
		if r.onUpdate != nil {
			r.onUpdate()
		}

		// Draw frame
		if err := r.drawFrame(); err != nil {
			return fmt.Errorf("draw frame failed: %w", err)
		}
	}

	vk.DeviceWaitIdle(r.device)
	return nil
}

func (r *VulkanRenderer) drawFrame() error {
	// Wait for previous frame
	fences := []vk.Fence{r.inFlightFence}
	vk.WaitForFences(r.device, 1, fences, vk.True, ^uint64(0))
	vk.ResetFences(r.device, 1, fences)

	// Acquire next image
	var imageIndex uint32
	result := vk.AcquireNextImage(r.device, r.swapchain, ^uint64(0), r.imageAvailableSem, vk.NullFence, &imageIndex)
	if result != vk.Success && result != vk.Suboptimal {
		return fmt.Errorf("failed to acquire swapchain image: %d", result)
	}

	// Reset and record command buffer
	vk.ResetCommandBuffer(r.commandBuffers[imageIndex], 0)
	if err := r.recordCommandBuffer(r.commandBuffers[imageIndex], imageIndex); err != nil {
		return err
	}

	// Submit command buffer
	waitSemaphores := []vk.Semaphore{r.imageAvailableSem}
	waitStages := []vk.PipelineStageFlags{vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)}
	signalSemaphores := []vk.Semaphore{r.renderFinishedSem}

	submitInfo := vk.SubmitInfo{
		SType:                vk.StructureTypeSubmitInfo,
		WaitSemaphoreCount:   1,
		PWaitSemaphores:      waitSemaphores,
		PWaitDstStageMask:    waitStages,
		CommandBufferCount:   1,
		PCommandBuffers:      []vk.CommandBuffer{r.commandBuffers[imageIndex]},
		SignalSemaphoreCount: 1,
		PSignalSemaphores:    signalSemaphores,
	}

	if result := vk.QueueSubmit(r.graphicsQueue, 1, []vk.SubmitInfo{submitInfo}, r.inFlightFence); result != vk.Success {
		return fmt.Errorf("failed to submit draw command buffer: %d", result)
	}

	// Present
	swapchains := []vk.Swapchain{r.swapchain}
	presentInfo := vk.PresentInfo{
		SType:              vk.StructureTypePresentInfo,
		WaitSemaphoreCount: 1,
		PWaitSemaphores:    signalSemaphores,
		SwapchainCount:     1,
		PSwapchains:        swapchains,
		PImageIndices:      []uint32{imageIndex},
	}

	vk.QueuePresent(r.presentQueue, &presentInfo)

	return nil
}

// Stop terminates the render loop.
func (r *VulkanRenderer) Stop() {
	r.running = false
}

// Cleanup releases all Vulkan resources.
func (r *VulkanRenderer) Cleanup() {
	if r.device != nil {
		vk.DeviceWaitIdle(r.device)

		// Destroy vertex buffer
		if r.vertexBuffer != nil {
			vk.DestroyBuffer(r.device, r.vertexBuffer, nil)
		}
		if r.vertexBufferMemory != nil {
			vk.FreeMemory(r.device, r.vertexBufferMemory, nil)
		}

		vk.DestroySemaphore(r.device, r.imageAvailableSem, nil)
		vk.DestroySemaphore(r.device, r.renderFinishedSem, nil)
		vk.DestroyFence(r.device, r.inFlightFence, nil)

		vk.DestroyCommandPool(r.device, r.commandPool, nil)

		for _, fb := range r.framebuffers {
			vk.DestroyFramebuffer(r.device, fb, nil)
		}

		if r.pipeline != nil {
			vk.DestroyPipeline(r.device, r.pipeline, nil)
		}
		if r.linePipeline != nil {
			vk.DestroyPipeline(r.device, r.linePipeline, nil)
		}
		vk.DestroyPipelineLayout(r.device, r.pipelineLayout, nil)
		vk.DestroyRenderPass(r.device, r.renderPass, nil)

		// Destroy shader modules
		if r.vertShaderModule != nil {
			vk.DestroyShaderModule(r.device, r.vertShaderModule, nil)
		}
		if r.fragShaderModule != nil {
			vk.DestroyShaderModule(r.device, r.fragShaderModule, nil)
		}

		for _, iv := range r.imageViews {
			vk.DestroyImageView(r.device, iv, nil)
		}

		vk.DestroySwapchain(r.device, r.swapchain, nil)
		vk.DestroyDevice(r.device, nil)
	}

	if r.instance != nil {
		vk.DestroySurface(r.instance, r.surface, nil)
		vk.DestroyInstance(r.instance, nil)
	}

	if r.window != nil {
		r.window.Destroy()
	}
	glfw.Terminate()
}

// Helper function for null-terminated strings
func safeString(s string) string {
	if len(s) == 0 || s[len(s)-1] != 0 {
		return s + "\x00"
	}
	return s
}
