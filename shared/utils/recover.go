// Package utils provides common utility functions for FeCIM tools
package utils

import (
	"fmt"
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

// SafeGoWithCallback runs a function in a goroutine with panic recovery.
// If the function panics, the onPanic callback is called with the panic value.
func SafeGoWithCallback(name string, fn func(), onPanic func(name string, panicValue interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if onPanic != nil {
					onPanic(name, r)
				} else {
					log.Printf("[PANIC] %s: %v\n%s", name, r, debug.Stack())
				}
			}
		}()
		fn()
	}()
}

// SafeCall runs a function with panic recovery and returns any error.
// If the function panics, the panic is converted to an error.
func SafeCall(name string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC] %s: %v\n%s", name, r, debug.Stack())
			err = fmt.Errorf("%s panicked: %v", name, r)
		}
	}()
	return fn()
}

// SafeClose closes a resource that implements Close(), recovering from any panic.
// Returns any error from Close(), or an error if Close() panicked.
func SafeClose(name string, closer interface{ Close() error }) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC] closing %s: %v", name, r)
		}
	}()
	if closer == nil {
		return nil
	}
	return closer.Close()
}
