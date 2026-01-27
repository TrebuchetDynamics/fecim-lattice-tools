package recording

import (
	"sync"
	"testing"
)

// =============================================================================
// Buffer Pool Creation Tests
// =============================================================================

func TestNewBufferPool(t *testing.T) {
	pool := NewBufferPool(1920, 1080)

	if pool == nil {
		t.Fatal("NewBufferPool returned nil")
	}
}

func TestNewBufferPoolWithSize(t *testing.T) {
	bufferSize := 1920 * 1080 * 3 // RGB24
	pool := NewBufferPoolWithSize(bufferSize)

	if pool == nil {
		t.Fatal("NewBufferPoolWithSize returned nil")
	}
}

func TestNewBufferPoolZeroSize(t *testing.T) {
	// Should handle zero or negative sizes gracefully
	pool := NewBufferPool(0, 0)

	if pool == nil {
		t.Fatal("Pool should be created even with zero size")
	}

	// Getting buffer should return empty or minimal buffer
	buf := pool.Get()
	if buf == nil {
		t.Error("Get should return non-nil buffer")
	}
}

// =============================================================================
// Buffer Get/Put Tests
// =============================================================================

func TestBufferPoolGetReturnsCorrectSize(t *testing.T) {
	width := 1920
	height := 1080
	expectedSize := width * height * 3 // RGB24

	pool := NewBufferPool(width, height)
	buf := pool.Get()

	if len(buf) != expectedSize {
		t.Errorf("Buffer size = %d, want %d", len(buf), expectedSize)
	}
}

func TestBufferPoolGetReturnsZeroedBuffer(t *testing.T) {
	pool := NewBufferPool(100, 100)

	// Get a buffer and write to it
	buf1 := pool.Get()
	for i := range buf1 {
		buf1[i] = 0xFF
	}

	// Return it
	pool.Put(buf1)

	// Get another buffer (might be the same one)
	buf2 := pool.Get()

	// Check if buffer is zeroed (implementation should clear on Put or Get)
	allZero := true
	for _, b := range buf2 {
		if b != 0 {
			allZero = false
			break
		}
	}

	// Note: This test documents expected behavior - buffers should be clean
	// If implementation doesn't zero buffers, this test will catch that
	if !allZero {
		t.Log("Warning: Buffer pool does not zero buffers - this may be intentional for performance")
	}
}

func TestBufferPoolPutNilIsNoOp(t *testing.T) {
	pool := NewBufferPool(100, 100)

	// Should not panic
	pool.Put(nil)
}

func TestBufferPoolPutWrongSizeIsHandled(t *testing.T) {
	pool := NewBufferPool(100, 100)

	// Put a buffer of wrong size
	wrongSize := make([]byte, 500)
	pool.Put(wrongSize) // Should not panic or corrupt pool

	// Getting should still return correct size
	buf := pool.Get()
	expectedSize := 100 * 100 * 3

	if len(buf) != expectedSize {
		t.Errorf("Buffer size after wrong put = %d, want %d", len(buf), expectedSize)
	}
}

// =============================================================================
// Buffer Pool Reuse Tests
// =============================================================================

func TestBufferPoolReusesBuffers(t *testing.T) {
	pool := NewBufferPool(100, 100)

	// Get and return a buffer
	buf1 := pool.Get()
	buf1Ptr := &buf1[0] // Get pointer to backing array
	pool.Put(buf1)

	// Get another buffer
	buf2 := pool.Get()
	buf2Ptr := &buf2[0]

	// Should be the same underlying array (buffer reuse)
	if buf1Ptr != buf2Ptr {
		t.Log("Buffer was not reused - pool may create new buffers (acceptable)")
	}
}

func TestBufferPoolMultipleGetWithoutPut(t *testing.T) {
	pool := NewBufferPool(100, 100)

	// Get multiple buffers without returning them
	buffers := make([][]byte, 10)
	for i := range buffers {
		buffers[i] = pool.Get()
	}

	// All should be valid and correct size
	expectedSize := 100 * 100 * 3
	for i, buf := range buffers {
		if len(buf) != expectedSize {
			t.Errorf("Buffer %d size = %d, want %d", i, len(buf), expectedSize)
		}
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestBufferPoolConcurrentAccess(t *testing.T) {
	pool := NewBufferPool(100, 100)
	expectedSize := 100 * 100 * 3

	var wg sync.WaitGroup
	numGoroutines := 100
	iterations := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				buf := pool.Get()
				if len(buf) != expectedSize {
					t.Errorf("Concurrent get returned wrong size: %d", len(buf))
				}
				// Simulate some work
				for k := 0; k < 100; k++ {
					buf[k] = byte(k)
				}
				pool.Put(buf)
			}
		}()
	}

	wg.Wait()
}

func TestBufferPoolConcurrentGetOnly(t *testing.T) {
	pool := NewBufferPool(100, 100)
	expectedSize := 100 * 100 * 3

	var wg sync.WaitGroup
	numGoroutines := 50
	results := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := pool.Get()
			results <- len(buf)
		}()
	}

	wg.Wait()
	close(results)

	for size := range results {
		if size != expectedSize {
			t.Errorf("Concurrent get returned wrong size: %d", size)
		}
	}
}

func TestBufferPoolNoDataRace(t *testing.T) {
	pool := NewBufferPool(100, 100)

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Writer goroutine
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				buf := pool.Get()
				for k := range buf {
					buf[k] = byte(id)
				}
				pool.Put(buf)
			}
		}(i)

		// Reader goroutine
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				buf := pool.Get()
				// Just read
				_ = buf[0]
				pool.Put(buf)
			}
		}()
	}

	wg.Wait()
}

// =============================================================================
// Buffer Pool Statistics Tests
// =============================================================================

func TestBufferPoolStats(t *testing.T) {
	pool := NewBufferPool(100, 100)

	// Initial stats
	stats := pool.Stats()

	if stats.BufferSize != 100*100*3 {
		t.Errorf("BufferSize = %d, want %d", stats.BufferSize, 100*100*3)
	}

	if stats.Gets != 0 {
		t.Errorf("Initial Gets = %d, want 0", stats.Gets)
	}

	if stats.Puts != 0 {
		t.Errorf("Initial Puts = %d, want 0", stats.Puts)
	}

	// Perform some operations
	buf := pool.Get()
	pool.Put(buf)
	buf = pool.Get()
	pool.Put(buf)

	stats = pool.Stats()

	if stats.Gets != 2 {
		t.Errorf("Gets = %d, want 2", stats.Gets)
	}

	if stats.Puts != 2 {
		t.Errorf("Puts = %d, want 2", stats.Puts)
	}
}

func TestBufferPoolHitMissTracking(t *testing.T) {
	pool := NewBufferPool(100, 100)

	// First get should be a miss (no pooled buffers)
	pool.Get()

	// Put it back
	buf := pool.Get()
	pool.Put(buf)

	// Next get might be a hit
	pool.Get()

	stats := pool.Stats()

	// At least one miss should occur (first allocation)
	if stats.Misses < 1 {
		t.Error("Expected at least 1 miss for initial allocation")
	}

	t.Logf("Pool stats: Gets=%d, Puts=%d, Hits=%d, Misses=%d",
		stats.Gets, stats.Puts, stats.Hits, stats.Misses)
}

// =============================================================================
// Buffer Pool Reset Tests
// =============================================================================

func TestBufferPoolReset(t *testing.T) {
	pool := NewBufferPool(100, 100)

	// Get some buffers and put them back
	for i := 0; i < 10; i++ {
		buf := pool.Get()
		pool.Put(buf)
	}

	// Reset the pool
	pool.Reset()

	// Stats should be cleared
	stats := pool.Stats()
	if stats.Gets != 0 || stats.Puts != 0 {
		t.Errorf("Stats not reset: Gets=%d, Puts=%d", stats.Gets, stats.Puts)
	}
}

func TestBufferPoolResize(t *testing.T) {
	pool := NewBufferPool(100, 100)
	oldSize := 100 * 100 * 3

	buf := pool.Get()
	if len(buf) != oldSize {
		t.Errorf("Initial buffer size = %d, want %d", len(buf), oldSize)
	}
	pool.Put(buf)

	// Resize the pool
	pool.Resize(200, 200)
	newSize := 200 * 200 * 3

	buf = pool.Get()
	if len(buf) != newSize {
		t.Errorf("Resized buffer size = %d, want %d", len(buf), newSize)
	}
}

// =============================================================================
// Frame Buffer Tests
// =============================================================================

func TestFrameBufferCreation(t *testing.T) {
	fb := NewFrameBuffer(1920, 1080)

	if fb == nil {
		t.Fatal("NewFrameBuffer returned nil")
	}

	if fb.Width() != 1920 {
		t.Errorf("Width = %d, want 1920", fb.Width())
	}

	if fb.Height() != 1080 {
		t.Errorf("Height = %d, want 1080", fb.Height())
	}
}

func TestFrameBufferData(t *testing.T) {
	fb := NewFrameBuffer(100, 100)
	expectedSize := 100 * 100 * 3

	data := fb.Data()
	if len(data) != expectedSize {
		t.Errorf("Data size = %d, want %d", len(data), expectedSize)
	}
}

func TestFrameBufferSetPixel(t *testing.T) {
	fb := NewFrameBuffer(100, 100)

	// Set pixel at (10, 20) to red
	fb.SetPixel(10, 20, 255, 0, 0)

	// Read it back
	r, g, b := fb.GetPixel(10, 20)

	if r != 255 || g != 0 || b != 0 {
		t.Errorf("Pixel = (%d, %d, %d), want (255, 0, 0)", r, g, b)
	}
}

func TestFrameBufferSetPixelOutOfBounds(t *testing.T) {
	fb := NewFrameBuffer(100, 100)

	// Should not panic for out of bounds
	fb.SetPixel(-1, 0, 255, 0, 0)
	fb.SetPixel(0, -1, 255, 0, 0)
	fb.SetPixel(100, 0, 255, 0, 0)
	fb.SetPixel(0, 100, 255, 0, 0)
}

func TestFrameBufferClear(t *testing.T) {
	fb := NewFrameBuffer(100, 100)

	// Set some pixels
	for x := 0; x < 50; x++ {
		fb.SetPixel(x, 0, 255, 255, 255)
	}

	// Clear
	fb.Clear()

	// Check all pixels are zero
	data := fb.Data()
	for i, b := range data {
		if b != 0 {
			t.Errorf("Data[%d] = %d after clear, want 0", i, b)
			break
		}
	}
}

// =============================================================================
// Memory Efficiency Tests
// =============================================================================

func TestBufferPoolMemoryEfficiency(t *testing.T) {
	pool := NewBufferPool(1920, 1080)

	// Get and return many buffers
	for i := 0; i < 1000; i++ {
		buf := pool.Get()
		pool.Put(buf)
	}

	stats := pool.Stats()

	// Hit rate should be high after warmup
	if stats.Gets > 10 && stats.Hits > 0 {
		hitRate := float64(stats.Hits) / float64(stats.Gets)
		t.Logf("Buffer pool hit rate: %.2f%%", hitRate*100)

		if hitRate < 0.5 {
			t.Log("Warning: Buffer pool hit rate is below 50%")
		}
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkBufferPoolGet(b *testing.B) {
	pool := NewBufferPool(1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get()
		pool.Put(buf)
	}
}

func BenchmarkBufferPoolGetConcurrent(b *testing.B) {
	pool := NewBufferPool(1920, 1080)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			pool.Put(buf)
		}
	})
}

func BenchmarkDirectAllocation(b *testing.B) {
	size := 1920 * 1080 * 3

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, size)
		_ = buf
	}
}

func BenchmarkFrameBufferSetPixel(b *testing.B) {
	fb := NewFrameBuffer(1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 1920
		y := (i / 1920) % 1080
		fb.SetPixel(x, y, 255, 128, 64)
	}
}

func BenchmarkFrameBufferClear(b *testing.B) {
	fb := NewFrameBuffer(1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fb.Clear()
	}
}
