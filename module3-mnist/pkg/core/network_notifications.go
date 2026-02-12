package core

// SetNotificationHandler registers an optional callback for user-facing notices
// (e.g., non-fatal fallback/clamping events).
func (net *DualModeNetwork) SetNotificationHandler(handler func(message string)) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.notifyUser = handler
}

func (net *DualModeNetwork) emitNotification(message string) {
	net.mu.RLock()
	handler := net.notifyUser
	net.mu.RUnlock()
	if handler != nil {
		handler(message)
	}
}
