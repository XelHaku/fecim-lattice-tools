// Package render provides Vulkan-based visualization for ferroelectric simulations.
package render

import (
	"fmt"

	"github.com/go-gl/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

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
	framebuffers   []vk.Framebuffer

	// Commands
	commandPool    vk.CommandPool
	commandBuffers []vk.CommandBuffer

	// Sync
	imageAvailableSem vk.Semaphore
	renderFinishedSem vk.Semaphore
	inFlightFence     vk.Fence

	// Callbacks
	onUpdate func()
	onKey    func(key glfw.Key, action glfw.Action)

	// Queue family indices
	graphicsFamily uint32
	presentFamily  uint32
}

// NewVulkanRenderer creates a new Vulkan-based renderer.
func NewVulkanRenderer(config *Config) *VulkanRenderer {
	return &VulkanRenderer{
		config: config,
		cell:   NewCellDisplay(),
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

// SetKeyCallback sets a function to be called on key events.
func (r *VulkanRenderer) SetKeyCallback(fn func(key glfw.Key, action glfw.Action)) {
	r.onKey = fn
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

	// Set up key callback
	r.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if r.onKey != nil {
			r.onKey(key, action)
		}
	})

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
	// For now, create a minimal pipeline layout without actual shaders
	// Real implementation would load compiled SPIR-V shaders

	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{
		SType: vk.StructureTypePipelineLayoutCreateInfo,
	}

	var pipelineLayout vk.PipelineLayout
	if result := vk.CreatePipelineLayout(r.device, &pipelineLayoutInfo, nil, &pipelineLayout); result != vk.Success {
		return fmt.Errorf("failed to create pipeline layout: %d", result)
	}
	r.pipelineLayout = pipelineLayout

	// Note: Full pipeline creation requires shader modules
	// For now we'll skip the pipeline and use a clear-only render pass

	return nil
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

func (r *VulkanRenderer) recordCommandBuffer(commandBuffer vk.CommandBuffer, imageIndex uint32) error {
	beginInfo := vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
	}

	if result := vk.BeginCommandBuffer(commandBuffer, &beginInfo); result != vk.Success {
		return fmt.Errorf("failed to begin command buffer: %d", result)
	}

	// Clear color based on polarization - creates a dynamic background
	normP := r.cell.Polarization
	var clearR, clearG, clearB float32 = 0.1, 0.1, 0.15

	// Shift background color slightly based on polarization
	if normP > 0 {
		clearR += float32(normP) * 0.1
	} else {
		clearB += float32(-normP) * 0.1
	}

	clearValue := vk.ClearValue{}
	clearValue.SetColor([]float32{clearR, clearG, clearB, 1.0})

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

	// Without a proper graphics pipeline, we can only clear the screen
	// The clear color changes based on polarization state

	vk.CmdEndRenderPass(commandBuffer)

	if result := vk.EndCommandBuffer(commandBuffer); result != vk.Success {
		return fmt.Errorf("failed to end command buffer: %d", result)
	}

	return nil
}

// Run starts the main render loop.
func (r *VulkanRenderer) Run() error {
	r.running = true

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
		vk.DestroyPipelineLayout(r.device, r.pipelineLayout, nil)
		vk.DestroyRenderPass(r.device, r.renderPass, nil)

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
