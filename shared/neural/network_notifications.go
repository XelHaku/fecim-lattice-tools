package neural

// SetNotificationHandler registers an optional callback for user-facing notices.
func (net *DualModeNetwork) SetNotificationHandler(handler func(message string)) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.notifyUser = handler
}

func (net *DualModeNetwork) emitNotification(message string) {
	net.mu.RLock()
	h := net.notifyUser
	net.mu.RUnlock()
	if h != nil {
		h(message)
	}
}
