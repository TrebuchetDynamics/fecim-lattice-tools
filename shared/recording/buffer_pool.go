package recording

import (
	"sync"
	"sync/atomic"
)

// BufferPool provides efficient buffer allocation and reuse for frame data.
type BufferPool struct {
	pool       sync.Pool
	mu         sync.RWMutex // Protects bufferSize access
	bufferSize int
	stats      bufferPoolStatsInternal
}

type bufferPoolStatsInternal struct {
	gets   int64
	puts   int64
	hits   int64
	misses int64
}

// NewBufferPool creates a new buffer pool for the given dimensions.
func NewBufferPool(width, height int) *BufferPool {
	size := width * height * 3 // RGB24
	if size <= 0 {
		size = 1 // Minimum size to avoid issues
	}
	return NewBufferPoolWithSize(size)
}

// NewBufferPoolWithSize creates a new buffer pool with a specific buffer size.
func NewBufferPoolWithSize(size int) *BufferPool {
	if size <= 0 {
		size = 1
	}

	bp := &BufferPool{
		bufferSize: size,
	}

	bp.pool.New = func() interface{} {
		atomic.AddInt64(&bp.stats.misses, 1)
		return make([]byte, bp.bufferSize)
	}

	return bp
}

// Get retrieves a buffer from the pool.
func (bp *BufferPool) Get() []byte {
	atomic.AddInt64(&bp.stats.gets, 1)

	bp.mu.RLock()
	currentSize := bp.bufferSize
	bp.mu.RUnlock()

	buf := bp.pool.Get().([]byte)

	// Ensure correct size (in case of resize)
	if len(buf) != currentSize {
		return make([]byte, currentSize)
	}

	return buf
}

// Put returns a buffer to the pool.
func (bp *BufferPool) Put(buf []byte) {
	if buf == nil {
		return
	}

	atomic.AddInt64(&bp.stats.puts, 1)

	bp.mu.RLock()
	currentSize := bp.bufferSize
	bp.mu.RUnlock()

	// Only return buffers of the correct size
	if len(buf) != currentSize {
		return
	}

	atomic.AddInt64(&bp.stats.hits, 1)
	bp.pool.Put(buf)
}

// Stats returns current pool statistics.
func (bp *BufferPool) Stats() BufferPoolStats {
	bp.mu.RLock()
	size := bp.bufferSize
	bp.mu.RUnlock()

	return BufferPoolStats{
		BufferSize: size,
		Gets:       atomic.LoadInt64(&bp.stats.gets),
		Puts:       atomic.LoadInt64(&bp.stats.puts),
		Hits:       atomic.LoadInt64(&bp.stats.hits),
		Misses:     atomic.LoadInt64(&bp.stats.misses),
	}
}

// Reset clears the pool statistics.
func (bp *BufferPool) Reset() {
	atomic.StoreInt64(&bp.stats.gets, 0)
	atomic.StoreInt64(&bp.stats.puts, 0)
	atomic.StoreInt64(&bp.stats.hits, 0)
	atomic.StoreInt64(&bp.stats.misses, 0)
}

// Resize changes the buffer size for future allocations.
// Thread-safe: uses mutex to protect bufferSize modification.
func (bp *BufferPool) Resize(width, height int) {
	newSize := width * height * 3
	if newSize <= 0 {
		newSize = 1
	}
	bp.mu.Lock()
	bp.bufferSize = newSize
	bp.mu.Unlock()
}

// =============================================================================
// Frame Buffer
// =============================================================================

// FrameBuffer represents a single video frame.
type FrameBuffer struct {
	data   []byte
	width  int
	height int
}

// NewFrameBuffer creates a new frame buffer.
func NewFrameBuffer(width, height int) *FrameBuffer {
	return &FrameBuffer{
		data:   make([]byte, width*height*3),
		width:  width,
		height: height,
	}
}

// Width returns the frame width.
func (fb *FrameBuffer) Width() int {
	return fb.width
}

// Height returns the frame height.
func (fb *FrameBuffer) Height() int {
	return fb.height
}

// Data returns the raw pixel data.
func (fb *FrameBuffer) Data() []byte {
	return fb.data
}

// SetPixel sets a pixel value (bounds-checked).
func (fb *FrameBuffer) SetPixel(x, y int, r, g, b byte) {
	if x < 0 || x >= fb.width || y < 0 || y >= fb.height {
		return
	}

	idx := (y*fb.width + x) * 3
	fb.data[idx] = r
	fb.data[idx+1] = g
	fb.data[idx+2] = b
}

// GetPixel gets a pixel value (bounds-checked).
func (fb *FrameBuffer) GetPixel(x, y int) (r, g, b byte) {
	if x < 0 || x >= fb.width || y < 0 || y >= fb.height {
		return 0, 0, 0
	}

	idx := (y*fb.width + x) * 3
	return fb.data[idx], fb.data[idx+1], fb.data[idx+2]
}

// Clear zeroes all pixel data.
func (fb *FrameBuffer) Clear() {
	for i := range fb.data {
		fb.data[i] = 0
	}
}
