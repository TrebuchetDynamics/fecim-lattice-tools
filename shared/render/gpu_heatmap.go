//go:build !js && cgo

// Package render provides GPU-accelerated rendering backends for crossbar
// array visualization. The GPUHeatmapRenderer uses headless Vulkan offscreen
// rendering with instanced vertex/fragment shaders to colour an NxN
// conductance matrix into an RGBA image, falling back gracefully when no GPU
// is available.
package render

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"sync"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// Embedded SPIR-V shader binaries for the heatmap graphics pipeline.
// Compiled from heatmap.vert and heatmap.frag in module1-hysteresis/shaders/.

//go:embed heatmap.vert.spv
var heatmapVertSPV []byte

//go:embed heatmap.frag.spv
var heatmapFragSPV []byte

// pushConstants matches the push constant layout in heatmap.vert.
// 6 fields x 4 bytes = 24 bytes total.
type pushConstants struct {
	Rows       uint32
	Cols       uint32
	OriginX    float32
	OriginY    float32
	CellWidth  float32
	CellHeight float32
}

const pushConstantsSize = 24

// GPUHeatmapRenderer renders crossbar heatmaps using Vulkan offscreen rendering.
//
// It creates an offscreen framebuffer, draws instanced triangle-strip quads
// (one per cell) coloured by a viridis polynomial in the fragment shader,
// then reads back pixels into an image.RGBA compatible with Fyne canvas.Raster.
//
// If Vulkan initialisation fails, Available() returns false and RenderHeatmap
// returns nil. The caller is expected to fall back to a software path.
type GPUHeatmapRenderer struct {
	mu        sync.Mutex
	available bool

	// Vulkan core (headless, no window/surface)
	instance       vk.Instance
	physicalDevice vk.PhysicalDevice
	device         vk.Device
	graphicsQueue  vk.Queue
	graphicsFamily uint32
	memoryProps    vk.PhysicalDeviceMemoryProperties

	// Graphics pipeline
	renderPass       vk.RenderPass
	descriptorLayout vk.DescriptorSetLayout
	pipelineLayout   vk.PipelineLayout
	pipeline         vk.Pipeline
	vertModule       vk.ShaderModule
	fragModule       vk.ShaderModule

	// Descriptor pool and set for the storage buffer
	descriptorPool vk.DescriptorPool
	descriptorSet  vk.DescriptorSet

	// Command infrastructure
	commandPool   vk.CommandPool
	commandBuffer vk.CommandBuffer
	fence         vk.Fence

	// Offscreen framebuffer (re-created on resolution change)
	fbImage     vk.Image
	fbMemory    vk.DeviceMemory
	fbView      vk.ImageView
	framebuffer vk.Framebuffer
	fbWidth     uint32
	fbHeight    uint32

	// Readback buffer (host-visible, for pixel download)
	readbackBuffer vk.Buffer
	readbackMemory vk.DeviceMemory
	readbackSize   vk.DeviceSize

	// Cell value storage buffer (host-visible for upload, binding 0 in shader)
	storageBuffer vk.Buffer
	storageMemory vk.DeviceMemory
	storageSize   vk.DeviceSize
	storageMapped unsafe.Pointer
}

// NewGPUHeatmapRenderer creates a GPU heatmap renderer.
// If Vulkan is not available, the renderer is created with Available() == false
// and all render calls return nil. Never panics on missing GPU.
func NewGPUHeatmapRenderer() *GPUHeatmapRenderer {
	r := &GPUHeatmapRenderer{}
	if err := r.init(); err != nil {
		r.available = false
	}
	return r
}

// Available returns whether GPU rendering is functional.
func (r *GPUHeatmapRenderer) Available() bool {
	return r.available
}

// RenderHeatmap renders a rows x cols heatmap with the given normalised values [0,1].
// values is row-major with length >= rows*cols.
// Returns an RGBA image of the specified pixel dimensions (pixW x pixH),
// or nil if the GPU is unavailable or inputs are invalid.
func (r *GPUHeatmapRenderer) RenderHeatmap(values []float64, rows, cols, pixW, pixH int) image.Image {
	if !r.available {
		return nil
	}
	if rows <= 0 || cols <= 0 || pixW <= 0 || pixH <= 0 {
		return nil
	}
	if len(values) < rows*cols {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure offscreen framebuffer matches requested resolution.
	if err := r.ensureFramebuffer(uint32(pixW), uint32(pixH)); err != nil {
		return nil
	}

	// Ensure storage buffer is large enough for cell data (float32 per cell).
	cellCount := rows * cols
	requiredStorage := vk.DeviceSize(cellCount * 4)
	if err := r.ensureStorageBuffer(requiredStorage); err != nil {
		return nil
	}

	// Upload cell values.
	r.uploadCellValues(values, rows*cols)

	// Update descriptor set to point to storage buffer.
	r.updateDescriptorSet()

	// Record and submit draw commands.
	if err := r.recordAndSubmit(uint32(rows), uint32(cols), uint32(pixW), uint32(pixH)); err != nil {
		return nil
	}

	// Read back pixels into image.RGBA.
	return r.readbackPixels(uint32(pixW), uint32(pixH))
}

// Destroy releases all Vulkan resources. Safe to call multiple times.
func (r *GPUHeatmapRenderer) Destroy() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.available && r.device == nil {
		return
	}

	if r.device != nil {
		vk.DeviceWaitIdle(r.device)
	}

	r.destroyStorageBuffer()
	r.destroyReadbackBuffer()
	r.destroyFramebuffer()

	if r.fence != nil {
		vk.DestroyFence(r.device, r.fence, nil)
		r.fence = nil
	}
	if r.commandPool != nil {
		vk.DestroyCommandPool(r.device, r.commandPool, nil)
		r.commandPool = nil
	}
	if r.descriptorPool != nil {
		vk.DestroyDescriptorPool(r.device, r.descriptorPool, nil)
		r.descriptorPool = nil
		r.descriptorSet = nil
	}
	if r.pipeline != nil {
		vk.DestroyPipeline(r.device, r.pipeline, nil)
		r.pipeline = nil
	}
	if r.pipelineLayout != nil {
		vk.DestroyPipelineLayout(r.device, r.pipelineLayout, nil)
		r.pipelineLayout = nil
	}
	if r.descriptorLayout != nil {
		vk.DestroyDescriptorSetLayout(r.device, r.descriptorLayout, nil)
		r.descriptorLayout = nil
	}
	if r.renderPass != nil {
		vk.DestroyRenderPass(r.device, r.renderPass, nil)
		r.renderPass = nil
	}
	if r.vertModule != nil {
		vk.DestroyShaderModule(r.device, r.vertModule, nil)
		r.vertModule = nil
	}
	if r.fragModule != nil {
		vk.DestroyShaderModule(r.device, r.fragModule, nil)
		r.fragModule = nil
	}
	if r.device != nil {
		vk.DestroyDevice(r.device, nil)
		r.device = nil
	}
	if r.instance != nil {
		vk.DestroyInstance(r.instance, nil)
		r.instance = nil
	}

	r.available = false
}

// ---------------------------------------------------------------------------
// Initialisation (headless Vulkan, no window)
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) init() error {
	if err := vk.SetDefaultGetInstanceProcAddr(); err != nil {
		return fmt.Errorf("vulkan loader: %w", err)
	}
	if err := vk.Init(); err != nil {
		return fmt.Errorf("vulkan init: %w", err)
	}

	if err := r.createInstance(); err != nil {
		return err
	}
	if err := r.pickPhysicalDevice(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.createLogicalDevice(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.createCommandPool(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.allocateCommandBuffer(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.createFence(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.createRenderPass(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.createShaderModules(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.createDescriptorSetLayout(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.createPipelineLayout(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.createGraphicsPipeline(); err != nil {
		r.Destroy()
		return err
	}
	if err := r.createDescriptorPoolAndSet(); err != nil {
		r.Destroy()
		return err
	}

	r.available = true
	return nil
}

func (r *GPUHeatmapRenderer) createInstance() error {
	appInfo := vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PApplicationName:   safeStr("FeCIM Heatmap"),
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        safeStr("FeCIM"),
		EngineVersion:      vk.MakeVersion(1, 0, 0),
		ApiVersion:         vk.ApiVersion11,
	}

	ci := vk.InstanceCreateInfo{
		SType:            vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo: &appInfo,
	}

	var inst vk.Instance
	if res := vk.CreateInstance(&ci, nil, &inst); res != vk.Success {
		return fmt.Errorf("vkCreateInstance: %d", res)
	}
	r.instance = inst
	vk.InitInstance(r.instance)
	return nil
}

func (r *GPUHeatmapRenderer) pickPhysicalDevice() error {
	var count uint32
	vk.EnumeratePhysicalDevices(r.instance, &count, nil)
	if count == 0 {
		return fmt.Errorf("no Vulkan GPU found")
	}

	devices := make([]vk.PhysicalDevice, count)
	vk.EnumeratePhysicalDevices(r.instance, &count, devices)

	for _, dev := range devices {
		if family, ok := r.findGraphicsFamily(dev); ok {
			r.physicalDevice = dev
			r.graphicsFamily = family
			vk.GetPhysicalDeviceMemoryProperties(dev, &r.memoryProps)
			r.memoryProps.Deref()
			return nil
		}
	}

	return fmt.Errorf("no GPU with graphics queue found")
}

func (r *GPUHeatmapRenderer) findGraphicsFamily(dev vk.PhysicalDevice) (uint32, bool) {
	var count uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(dev, &count, nil)
	families := make([]vk.QueueFamilyProperties, count)
	vk.GetPhysicalDeviceQueueFamilyProperties(dev, &count, families)

	for i, qf := range families {
		qf.Deref()
		if qf.QueueFlags&vk.QueueFlags(vk.QueueGraphicsBit) != 0 {
			return uint32(i), true
		}
	}
	return 0, false
}

func (r *GPUHeatmapRenderer) createLogicalDevice() error {
	priority := []float32{1.0}
	queueCI := vk.DeviceQueueCreateInfo{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: r.graphicsFamily,
		QueueCount:       1,
		PQueuePriorities: priority,
	}

	deviceCI := vk.DeviceCreateInfo{
		SType:                vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount: 1,
		PQueueCreateInfos:    []vk.DeviceQueueCreateInfo{queueCI},
	}

	var dev vk.Device
	if res := vk.CreateDevice(r.physicalDevice, &deviceCI, nil, &dev); res != vk.Success {
		return fmt.Errorf("vkCreateDevice: %d", res)
	}
	r.device = dev

	var queue vk.Queue
	vk.GetDeviceQueue(r.device, r.graphicsFamily, 0, &queue)
	r.graphicsQueue = queue
	return nil
}

func (r *GPUHeatmapRenderer) createCommandPool() error {
	poolCI := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: r.graphicsFamily,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
	}

	var pool vk.CommandPool
	if res := vk.CreateCommandPool(r.device, &poolCI, nil, &pool); res != vk.Success {
		return fmt.Errorf("vkCreateCommandPool: %d", res)
	}
	r.commandPool = pool
	return nil
}

func (r *GPUHeatmapRenderer) allocateCommandBuffer() error {
	allocInfo := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        r.commandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}

	bufs := make([]vk.CommandBuffer, 1)
	if res := vk.AllocateCommandBuffers(r.device, &allocInfo, bufs); res != vk.Success {
		return fmt.Errorf("vkAllocateCommandBuffers: %d", res)
	}
	r.commandBuffer = bufs[0]
	return nil
}

func (r *GPUHeatmapRenderer) createFence() error {
	fenceCI := vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
	}
	var fence vk.Fence
	if res := vk.CreateFence(r.device, &fenceCI, nil, &fence); res != vk.Success {
		return fmt.Errorf("vkCreateFence: %d", res)
	}
	r.fence = fence
	return nil
}

// ---------------------------------------------------------------------------
// Render pass (offscreen R8G8B8A8_UNORM, final layout = TRANSFER_SRC for readback)
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) createRenderPass() error {
	colorAttachment := vk.AttachmentDescription{
		Format:         vk.FormatR8g8b8a8Unorm,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutTransferSrcOptimal,
	}

	colorRef := vk.AttachmentReference{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}

	subpass := vk.SubpassDescription{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: 1,
		PColorAttachments:    []vk.AttachmentReference{colorRef},
	}

	// Dependency: colour writes must complete before transfer read.
	dep := vk.SubpassDependency{
		SrcSubpass:    0,
		DstSubpass:    vk.SubpassExternal,
		SrcStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		SrcAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
		DstStageMask:  vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit),
	}

	rpCI := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    []vk.AttachmentDescription{colorAttachment},
		SubpassCount:    1,
		PSubpasses:      []vk.SubpassDescription{subpass},
		DependencyCount: 1,
		PDependencies:   []vk.SubpassDependency{dep},
	}

	var rp vk.RenderPass
	if res := vk.CreateRenderPass(r.device, &rpCI, nil, &rp); res != vk.Success {
		return fmt.Errorf("vkCreateRenderPass: %d", res)
	}
	r.renderPass = rp
	return nil
}

// ---------------------------------------------------------------------------
// Shader modules (from go:embed SPIR-V)
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) createShaderModules() error {
	var err error
	r.vertModule, err = r.newShaderModule(heatmapVertSPV)
	if err != nil {
		return fmt.Errorf("vertex shader: %w", err)
	}
	r.fragModule, err = r.newShaderModule(heatmapFragSPV)
	if err != nil {
		return fmt.Errorf("fragment shader: %w", err)
	}
	return nil
}

func (r *GPUHeatmapRenderer) newShaderModule(spirv []byte) (vk.ShaderModule, error) {
	if len(spirv) == 0 || len(spirv)%4 != 0 {
		return nil, fmt.Errorf("invalid SPIR-V (len=%d)", len(spirv))
	}
	code := make([]uint32, len(spirv)/4)
	for i := range code {
		code[i] = binary.LittleEndian.Uint32(spirv[i*4:])
	}

	ci := vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len(spirv)),
		PCode:    code,
	}

	var mod vk.ShaderModule
	if res := vk.CreateShaderModule(r.device, &ci, nil, &mod); res != vk.Success {
		return nil, fmt.Errorf("vkCreateShaderModule: %d", res)
	}
	return mod, nil
}

// ---------------------------------------------------------------------------
// Descriptor set layout (binding 0 = storage buffer with cell values)
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) createDescriptorSetLayout() error {
	binding := vk.DescriptorSetLayoutBinding{
		Binding:         0,
		DescriptorType:  vk.DescriptorTypeStorageBuffer,
		DescriptorCount: 1,
		StageFlags:      vk.ShaderStageFlags(vk.ShaderStageVertexBit),
	}

	ci := vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: 1,
		PBindings:    []vk.DescriptorSetLayoutBinding{binding},
	}

	var layout vk.DescriptorSetLayout
	if res := vk.CreateDescriptorSetLayout(r.device, &ci, nil, &layout); res != vk.Success {
		return fmt.Errorf("vkCreateDescriptorSetLayout: %d", res)
	}
	r.descriptorLayout = layout
	return nil
}

// ---------------------------------------------------------------------------
// Pipeline layout (push constants + one descriptor set)
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) createPipelineLayout() error {
	pushRange := vk.PushConstantRange{
		StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		Offset:     0,
		Size:       pushConstantsSize,
	}

	ci := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{r.descriptorLayout},
		PushConstantRangeCount: 1,
		PPushConstantRanges:    []vk.PushConstantRange{pushRange},
	}

	var layout vk.PipelineLayout
	if res := vk.CreatePipelineLayout(r.device, &ci, nil, &layout); res != vk.Success {
		return fmt.Errorf("vkCreatePipelineLayout: %d", res)
	}
	r.pipelineLayout = layout
	return nil
}

// ---------------------------------------------------------------------------
// Graphics pipeline (instanced triangle strip, no vertex buffer)
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) createGraphicsPipeline() error {
	vertStage := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageVertexBit,
		Module: r.vertModule,
		PName:  safeStr("main"),
	}
	fragStage := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageFragmentBit,
		Module: r.fragModule,
		PName:  safeStr("main"),
	}
	stages := []vk.PipelineShaderStageCreateInfo{vertStage, fragStage}

	// No vertex input: positions computed from gl_VertexIndex and gl_InstanceIndex.
	vertexInput := vk.PipelineVertexInputStateCreateInfo{
		SType: vk.StructureTypePipelineVertexInputStateCreateInfo,
	}

	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               vk.PrimitiveTopologyTriangleStrip,
		PrimitiveRestartEnable: vk.False,
	}

	// Dynamic viewport and scissor for resolution-independent pipeline.
	dynStates := []vk.DynamicState{vk.DynamicStateViewport, vk.DynamicStateScissor}
	dynamicState := vk.PipelineDynamicStateCreateInfo{
		SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: uint32(len(dynStates)),
		PDynamicStates:    dynStates,
	}

	// Placeholder viewport/scissor (overridden by dynamic state at draw time).
	viewport := vk.Viewport{Width: 1, Height: 1, MaxDepth: 1.0}
	scissor := vk.Rect2D{Extent: vk.Extent2D{Width: 1, Height: 1}}
	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		PViewports:    []vk.Viewport{viewport},
		ScissorCount:  1,
		PScissors:     []vk.Rect2D{scissor},
	}

	rasterizer := vk.PipelineRasterizationStateCreateInfo{
		SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
		PolygonMode:             vk.PolygonModeFill,
		LineWidth:               1.0,
		CullMode:                vk.CullModeFlags(vk.CullModeNone),
		FrontFace:               vk.FrontFaceCounterClockwise,
		DepthClampEnable:        vk.False,
		RasterizerDiscardEnable: vk.False,
		DepthBiasEnable:         vk.False,
	}

	multisampling := vk.PipelineMultisampleStateCreateInfo{
		SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples: vk.SampleCount1Bit,
		SampleShadingEnable:  vk.False,
	}

	colorBlendAttachment := vk.PipelineColorBlendAttachmentState{
		ColorWriteMask: vk.ColorComponentFlags(
			vk.ColorComponentRBit | vk.ColorComponentGBit |
				vk.ColorComponentBBit | vk.ColorComponentABit),
		BlendEnable: vk.False,
	}

	colorBlending := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   vk.False,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{colorBlendAttachment},
	}

	pipelineCI := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          uint32(len(stages)),
		PStages:             stages,
		PVertexInputState:   &vertexInput,
		PInputAssemblyState: &inputAssembly,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterizer,
		PMultisampleState:   &multisampling,
		PColorBlendState:    &colorBlending,
		PDynamicState:       &dynamicState,
		Layout:              r.pipelineLayout,
		RenderPass:          r.renderPass,
		Subpass:             0,
	}

	pipelines := make([]vk.Pipeline, 1)
	if res := vk.CreateGraphicsPipelines(r.device, vk.NullPipelineCache, 1,
		[]vk.GraphicsPipelineCreateInfo{pipelineCI}, nil, pipelines); res != vk.Success {
		return fmt.Errorf("vkCreateGraphicsPipelines: %d", res)
	}
	r.pipeline = pipelines[0]
	return nil
}

// ---------------------------------------------------------------------------
// Descriptor pool and set
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) createDescriptorPoolAndSet() error {
	poolSize := vk.DescriptorPoolSize{
		Type:            vk.DescriptorTypeStorageBuffer,
		DescriptorCount: 1,
	}

	poolCI := vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		MaxSets:       1,
		PoolSizeCount: 1,
		PPoolSizes:    []vk.DescriptorPoolSize{poolSize},
	}

	var pool vk.DescriptorPool
	if res := vk.CreateDescriptorPool(r.device, &poolCI, nil, &pool); res != vk.Success {
		return fmt.Errorf("vkCreateDescriptorPool: %d", res)
	}
	r.descriptorPool = pool

	allocInfo := vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     r.descriptorPool,
		DescriptorSetCount: 1,
		PSetLayouts:        []vk.DescriptorSetLayout{r.descriptorLayout},
	}

	var set vk.DescriptorSet
	if res := vk.AllocateDescriptorSets(r.device, &allocInfo, &set); res != vk.Success {
		return fmt.Errorf("vkAllocateDescriptorSets: %d", res)
	}
	r.descriptorSet = set
	return nil
}

// ---------------------------------------------------------------------------
// Offscreen framebuffer management
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) ensureFramebuffer(w, h uint32) error {
	if r.fbWidth == w && r.fbHeight == h && r.framebuffer != nil {
		return nil
	}

	r.destroyFramebuffer()

	// Create offscreen image (RGBA8 UNORM, colour attachment + transfer src).
	imgCI := vk.ImageCreateInfo{
		SType:       vk.StructureTypeImageCreateInfo,
		ImageType:   vk.ImageType2d,
		Format:      vk.FormatR8g8b8a8Unorm,
		Extent:      vk.Extent3D{Width: w, Height: h, Depth: 1},
		MipLevels:   1,
		ArrayLayers: 1,
		Samples:     vk.SampleCount1Bit,
		Tiling:      vk.ImageTilingOptimal,
		Usage: vk.ImageUsageFlags(
			vk.ImageUsageColorAttachmentBit | vk.ImageUsageTransferSrcBit),
		SharingMode:   vk.SharingModeExclusive,
		InitialLayout: vk.ImageLayoutUndefined,
	}

	var img vk.Image
	if res := vk.CreateImage(r.device, &imgCI, nil, &img); res != vk.Success {
		return fmt.Errorf("vkCreateImage: %d", res)
	}
	r.fbImage = img

	// Allocate device-local memory.
	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(r.device, r.fbImage, &memReqs)
	memReqs.Deref()

	memIdx, err := r.findMemoryType(memReqs.MemoryTypeBits,
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit))
	if err != nil {
		return err
	}

	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memIdx,
	}

	var mem vk.DeviceMemory
	if res := vk.AllocateMemory(r.device, &allocInfo, nil, &mem); res != vk.Success {
		return fmt.Errorf("vkAllocateMemory (fb): %d", res)
	}
	r.fbMemory = mem

	if res := vk.BindImageMemory(r.device, r.fbImage, r.fbMemory, 0); res != vk.Success {
		return fmt.Errorf("vkBindImageMemory: %d", res)
	}

	// Create image view.
	viewCI := vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    r.fbImage,
		ViewType: vk.ImageViewType2d,
		Format:   vk.FormatR8g8b8a8Unorm,
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

	var view vk.ImageView
	if res := vk.CreateImageView(r.device, &viewCI, nil, &view); res != vk.Success {
		return fmt.Errorf("vkCreateImageView: %d", res)
	}
	r.fbView = view

	// Create framebuffer.
	fbCI := vk.FramebufferCreateInfo{
		SType:           vk.StructureTypeFramebufferCreateInfo,
		RenderPass:      r.renderPass,
		AttachmentCount: 1,
		PAttachments:    []vk.ImageView{r.fbView},
		Width:           w,
		Height:          h,
		Layers:          1,
	}

	var fb vk.Framebuffer
	if res := vk.CreateFramebuffer(r.device, &fbCI, nil, &fb); res != vk.Success {
		return fmt.Errorf("vkCreateFramebuffer: %d", res)
	}
	r.framebuffer = fb
	r.fbWidth = w
	r.fbHeight = h

	// Ensure readback buffer is large enough (4 bytes per pixel).
	return r.ensureReadbackBuffer(vk.DeviceSize(w) * vk.DeviceSize(h) * 4)
}

func (r *GPUHeatmapRenderer) destroyFramebuffer() {
	if r.device == nil {
		return
	}
	vk.DeviceWaitIdle(r.device)

	if r.framebuffer != nil {
		vk.DestroyFramebuffer(r.device, r.framebuffer, nil)
		r.framebuffer = nil
	}
	if r.fbView != nil {
		vk.DestroyImageView(r.device, r.fbView, nil)
		r.fbView = nil
	}
	if r.fbImage != nil {
		vk.DestroyImage(r.device, r.fbImage, nil)
		r.fbImage = nil
	}
	if r.fbMemory != nil {
		vk.FreeMemory(r.device, r.fbMemory, nil)
		r.fbMemory = nil
	}
	r.fbWidth = 0
	r.fbHeight = 0
}

// ---------------------------------------------------------------------------
// Readback buffer (host-visible, for downloading framebuffer pixels)
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) ensureReadbackBuffer(requiredSize vk.DeviceSize) error {
	if r.readbackSize >= requiredSize && r.readbackBuffer != nil {
		return nil
	}
	r.destroyReadbackBuffer()

	bufCI := vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        requiredSize,
		Usage:       vk.BufferUsageFlags(vk.BufferUsageTransferDstBit),
		SharingMode: vk.SharingModeExclusive,
	}

	var buf vk.Buffer
	if res := vk.CreateBuffer(r.device, &bufCI, nil, &buf); res != vk.Success {
		return fmt.Errorf("vkCreateBuffer (readback): %d", res)
	}
	r.readbackBuffer = buf

	var memReqs vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(r.device, r.readbackBuffer, &memReqs)
	memReqs.Deref()

	memIdx, err := r.findMemoryType(memReqs.MemoryTypeBits,
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memIdx,
	}

	var mem vk.DeviceMemory
	if res := vk.AllocateMemory(r.device, &allocInfo, nil, &mem); res != vk.Success {
		return fmt.Errorf("vkAllocateMemory (readback): %d", res)
	}
	r.readbackMemory = mem

	if res := vk.BindBufferMemory(r.device, r.readbackBuffer, r.readbackMemory, 0); res != vk.Success {
		return fmt.Errorf("vkBindBufferMemory (readback): %d", res)
	}
	r.readbackSize = requiredSize
	return nil
}

func (r *GPUHeatmapRenderer) destroyReadbackBuffer() {
	if r.device == nil {
		return
	}
	if r.readbackBuffer != nil {
		vk.DestroyBuffer(r.device, r.readbackBuffer, nil)
		r.readbackBuffer = nil
	}
	if r.readbackMemory != nil {
		vk.FreeMemory(r.device, r.readbackMemory, nil)
		r.readbackMemory = nil
	}
	r.readbackSize = 0
}

// ---------------------------------------------------------------------------
// Storage buffer (cell values, host-visible for upload)
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) ensureStorageBuffer(requiredSize vk.DeviceSize) error {
	// Minimum 16 bytes to avoid zero-size edge cases.
	if requiredSize < 16 {
		requiredSize = 16
	}
	if r.storageSize >= requiredSize && r.storageBuffer != nil {
		return nil
	}
	r.destroyStorageBuffer()

	bufCI := vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        requiredSize,
		Usage:       vk.BufferUsageFlags(vk.BufferUsageStorageBufferBit),
		SharingMode: vk.SharingModeExclusive,
	}

	var buf vk.Buffer
	if res := vk.CreateBuffer(r.device, &bufCI, nil, &buf); res != vk.Success {
		return fmt.Errorf("vkCreateBuffer (storage): %d", res)
	}
	r.storageBuffer = buf

	var memReqs vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(r.device, r.storageBuffer, &memReqs)
	memReqs.Deref()

	memIdx, err := r.findMemoryType(memReqs.MemoryTypeBits,
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memIdx,
	}

	var mem vk.DeviceMemory
	if res := vk.AllocateMemory(r.device, &allocInfo, nil, &mem); res != vk.Success {
		return fmt.Errorf("vkAllocateMemory (storage): %d", res)
	}
	r.storageMemory = mem

	if res := vk.BindBufferMemory(r.device, r.storageBuffer, r.storageMemory, 0); res != vk.Success {
		return fmt.Errorf("vkBindBufferMemory (storage): %d", res)
	}

	// Persistently map for CPU writes.
	var mapped unsafe.Pointer
	if res := vk.MapMemory(r.device, r.storageMemory, 0, requiredSize, 0, &mapped); res != vk.Success {
		return fmt.Errorf("vkMapMemory (storage): %d", res)
	}
	r.storageMapped = mapped
	r.storageSize = requiredSize
	return nil
}

func (r *GPUHeatmapRenderer) destroyStorageBuffer() {
	if r.device == nil {
		return
	}
	if r.storageMapped != nil {
		vk.UnmapMemory(r.device, r.storageMemory)
		r.storageMapped = nil
	}
	if r.storageBuffer != nil {
		vk.DestroyBuffer(r.device, r.storageBuffer, nil)
		r.storageBuffer = nil
	}
	if r.storageMemory != nil {
		vk.FreeMemory(r.device, r.storageMemory, nil)
		r.storageMemory = nil
	}
	r.storageSize = 0
}

// ---------------------------------------------------------------------------
// Upload cell values as float32 to the storage buffer
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) uploadCellValues(values []float64, n int) {
	dst := unsafe.Slice((*float32)(r.storageMapped), n)
	for i := 0; i < n; i++ {
		v := values[i]
		// Clamp to [0,1].
		if v < 0 {
			v = 0
		} else if v > 1 {
			v = 1
		}
		dst[i] = float32(v)
	}
}

// ---------------------------------------------------------------------------
// Descriptor set update
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) updateDescriptorSet() {
	bufInfo := vk.DescriptorBufferInfo{
		Buffer: r.storageBuffer,
		Offset: 0,
		Range:  r.storageSize,
	}

	write := vk.WriteDescriptorSet{
		SType:           vk.StructureTypeWriteDescriptorSet,
		DstSet:          r.descriptorSet,
		DstBinding:      0,
		DstArrayElement: 0,
		DescriptorCount: 1,
		DescriptorType:  vk.DescriptorTypeStorageBuffer,
		PBufferInfo:     []vk.DescriptorBufferInfo{bufInfo},
	}

	vk.UpdateDescriptorSets(r.device, 1, []vk.WriteDescriptorSet{write}, 0, nil)
}

// ---------------------------------------------------------------------------
// Record command buffer, submit, and wait
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) recordAndSubmit(rows, cols, pixW, pixH uint32) error {
	if res := vk.ResetCommandBuffer(r.commandBuffer, 0); res != vk.Success {
		return fmt.Errorf("vkResetCommandBuffer: %d", res)
	}

	beginInfo := vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	}
	if res := vk.BeginCommandBuffer(r.commandBuffer, &beginInfo); res != vk.Success {
		return fmt.Errorf("vkBeginCommandBuffer: %d", res)
	}

	// Begin render pass with dark background matching the software renderer.
	clearValue := vk.NewClearValue([]float32{0.118, 0.118, 0.157, 1.0})
	rpBeginInfo := vk.RenderPassBeginInfo{
		SType:           vk.StructureTypeRenderPassBeginInfo,
		RenderPass:      r.renderPass,
		Framebuffer:     r.framebuffer,
		RenderArea:      vk.Rect2D{Extent: vk.Extent2D{Width: pixW, Height: pixH}},
		ClearValueCount: 1,
		PClearValues:    []vk.ClearValue{clearValue},
	}

	vk.CmdBeginRenderPass(r.commandBuffer, &rpBeginInfo, vk.SubpassContentsInline)

	// Bind pipeline.
	vk.CmdBindPipeline(r.commandBuffer, vk.PipelineBindPointGraphics, r.pipeline)

	// Set dynamic viewport and scissor.
	viewport := vk.Viewport{
		X: 0, Y: 0,
		Width: float32(pixW), Height: float32(pixH),
		MinDepth: 0.0, MaxDepth: 1.0,
	}
	vk.CmdSetViewport(r.commandBuffer, 0, 1, []vk.Viewport{viewport})

	scissor := vk.Rect2D{Extent: vk.Extent2D{Width: pixW, Height: pixH}}
	vk.CmdSetScissor(r.commandBuffer, 0, 1, []vk.Rect2D{scissor})

	// Bind descriptor set.
	vk.CmdBindDescriptorSets(r.commandBuffer, vk.PipelineBindPointGraphics,
		r.pipelineLayout, 0, 1, []vk.DescriptorSet{r.descriptorSet}, 0, nil)

	// Calculate push constants: map grid to NDC with a 20-pixel margin.
	// NDC ranges from -1 to +1 in both axes.
	marginPixels := float64(20)
	marginNDCx := 2.0 * marginPixels / float64(pixW)
	marginNDCy := 2.0 * marginPixels / float64(pixH)
	gridNDCw := 2.0 - 2.0*marginNDCx
	gridNDCh := 2.0 - 2.0*marginNDCy

	// Square cells that fit the available area.
	cellNDCw := gridNDCw / float64(cols)
	cellNDCh := gridNDCh / float64(rows)
	cellNDC := math.Min(cellNDCw, cellNDCh)

	originX := -1.0 + marginNDCx
	originY := -1.0 + marginNDCy

	pc := pushConstants{
		Rows:       rows,
		Cols:       cols,
		OriginX:    float32(originX),
		OriginY:    float32(originY),
		CellWidth:  float32(cellNDC),
		CellHeight: float32(cellNDC),
	}

	// Serialise push constants to bytes.
	pcBytes := make([]byte, pushConstantsSize)
	binary.LittleEndian.PutUint32(pcBytes[0:4], pc.Rows)
	binary.LittleEndian.PutUint32(pcBytes[4:8], pc.Cols)
	binary.LittleEndian.PutUint32(pcBytes[8:12], math.Float32bits(pc.OriginX))
	binary.LittleEndian.PutUint32(pcBytes[12:16], math.Float32bits(pc.OriginY))
	binary.LittleEndian.PutUint32(pcBytes[16:20], math.Float32bits(pc.CellWidth))
	binary.LittleEndian.PutUint32(pcBytes[20:24], math.Float32bits(pc.CellHeight))

	vk.CmdPushConstants(r.commandBuffer, r.pipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageVertexBit), 0, uint32(len(pcBytes)),
		unsafe.Pointer(&pcBytes[0]))

	// Draw: 4 vertices (triangle strip quad) x (rows*cols) instances.
	instanceCount := rows * cols
	vk.CmdDraw(r.commandBuffer, 4, instanceCount, 0, 0)

	vk.CmdEndRenderPass(r.commandBuffer)

	// Copy rendered image to readback buffer.
	copyRegion := vk.BufferImageCopy{
		BufferOffset:      0,
		BufferRowLength:   0, // tightly packed
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageOffset: vk.Offset3D{},
		ImageExtent: vk.Extent3D{Width: pixW, Height: pixH, Depth: 1},
	}

	vk.CmdCopyImageToBuffer(r.commandBuffer, r.fbImage,
		vk.ImageLayoutTransferSrcOptimal, r.readbackBuffer,
		1, []vk.BufferImageCopy{copyRegion})

	if res := vk.EndCommandBuffer(r.commandBuffer); res != vk.Success {
		return fmt.Errorf("vkEndCommandBuffer: %d", res)
	}

	// Submit and wait.
	fences := []vk.Fence{r.fence}
	vk.ResetFences(r.device, 1, fences)

	submitInfo := vk.SubmitInfo{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{r.commandBuffer},
	}

	if res := vk.QueueSubmit(r.graphicsQueue, 1, []vk.SubmitInfo{submitInfo}, r.fence); res != vk.Success {
		return fmt.Errorf("vkQueueSubmit: %d", res)
	}

	if res := vk.WaitForFences(r.device, 1, fences, vk.True, ^uint64(0)); res != vk.Success {
		return fmt.Errorf("vkWaitForFences: %d", res)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Readback pixels
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) readbackPixels(w, h uint32) image.Image {
	byteSize := w * h * 4

	var mapped unsafe.Pointer
	if res := vk.MapMemory(r.device, r.readbackMemory, 0, vk.DeviceSize(byteSize), 0, &mapped); res != vk.Success {
		return nil
	}
	defer vk.UnmapMemory(r.device, r.readbackMemory)

	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	src := unsafe.Slice((*byte)(mapped), byteSize)
	copy(img.Pix, src)

	return img
}

// ---------------------------------------------------------------------------
// Memory type finder
// ---------------------------------------------------------------------------

func (r *GPUHeatmapRenderer) findMemoryType(typeBits uint32, props vk.MemoryPropertyFlags) (uint32, error) {
	for i := uint32(0); i < r.memoryProps.MemoryTypeCount; i++ {
		r.memoryProps.MemoryTypes[i].Deref()
		if (typeBits&(1<<i)) != 0 &&
			(r.memoryProps.MemoryTypes[i].PropertyFlags&props) == props {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no suitable memory type (bits=0x%x, props=0x%x)", typeBits, props)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// safeStr ensures a null-terminated string for Vulkan API calls.
func safeStr(s string) string {
	if len(s) == 0 || s[len(s)-1] != 0 {
		return s + "\x00"
	}
	return s
}
