// Package utils provides common utility functions for FeCIM tools
package utils

import (
	"log"
	"runtime/debug"
)

// SafeGo runs a function in a goroutine with panic recovery.
// If the function panics, the panic is logged with a stack trace
// but does not crash the application.
func SafeGo(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %s: %v\n%s", name, r, debug.Stack())
			}
		}()
		fn()
	}()
}
