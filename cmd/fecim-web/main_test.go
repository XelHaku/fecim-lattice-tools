package main

import "testing"

func TestWebEntrypointDefaults(t *testing.T) {
	if webWindowTitle == "" {
		t.Fatal("webWindowTitle must be non-empty")
	}
	if webWindowW <= 0 || webWindowH <= 0 {
		t.Fatalf("invalid default web window size: %.1fx%.1f", webWindowW, webWindowH)
	}
}
