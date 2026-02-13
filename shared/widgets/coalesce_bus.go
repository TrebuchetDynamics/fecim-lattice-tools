package widgets

import (
	"sync"
	"time"
)

// CoalesceBus debounces bursty update requests and only executes the last update per key.
type CoalesceBus struct {
	window  time.Duration
	mu      sync.Mutex
	pending map[string]*time.Timer
	closed  bool
}

// NewCoalesceBus creates a bus with the debounce window (recommended 30-50ms).
func NewCoalesceBus(window time.Duration) *CoalesceBus {
	if window <= 0 {
		window = 40 * time.Millisecond
	}
	return &CoalesceBus{window: window, pending: make(map[string]*time.Timer)}
}

// Submit debounces updates by key. Within the debounce window only the latest callback runs.
func (b *CoalesceBus) Submit(key string, fn func()) {
	if fn == nil {
		return
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	if existing := b.pending[key]; existing != nil {
		existing.Stop()
	}
	b.pending[key] = time.AfterFunc(b.window, func() {
		fn()
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.pending, key)
	})
	b.mu.Unlock()
}

// Close cancels pending work and rejects new submissions.
func (b *CoalesceBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	for _, timer := range b.pending {
		timer.Stop()
	}
	b.pending = map[string]*time.Timer{}
	b.closed = true
}
