package main

import (
	"fmt"
	"strings"
)

// HeadlessEngineOptions carries command-level options for a headless module run.
type HeadlessEngineOptions struct {
	Engine string
}

// HeadlessEngineReceipt summarizes the selected headless engine run.
type HeadlessEngineReceipt struct {
	Mode   string
	Engine string
}

// HeadlessEnginePort is the command-level seam for module headless runners.
type HeadlessEnginePort interface {
	Run(HeadlessEngineOptions) (HeadlessEngineReceipt, error)
}

type headlessEngineFunc func(HeadlessEngineOptions) (HeadlessEngineReceipt, error)

func (f headlessEngineFunc) Run(options HeadlessEngineOptions) (HeadlessEngineReceipt, error) {
	return f(options)
}

// HeadlessEngineEntry binds a module mode name to a headless engine port.
type HeadlessEngineEntry struct {
	Mode string
	Port HeadlessEnginePort
}

// HeadlessEngineRegistry returns command-level headless engines.
func HeadlessEngineRegistry() []HeadlessEngineEntry {
	return []HeadlessEngineEntry{{
		Mode: "hysteresis",
		Port: headlessEngineFunc(func(options HeadlessEngineOptions) (HeadlessEngineReceipt, error) {
			err := runHysteresisMode(options.Engine)
			return HeadlessEngineReceipt{Mode: "hysteresis", Engine: normalizeEngine(options.Engine)}, err
		}),
	}}
}

func runHeadlessEngine(mode string, options HeadlessEngineOptions) error {
	for _, entry := range HeadlessEngineRegistry() {
		if entry.Mode == mode {
			_, err := entry.Port.Run(options)
			return err
		}
	}
	return fmt.Errorf("unknown mode %q (expected: %s)", mode, knownHeadlessEngineModes())
}

func knownHeadlessEngineModes() string {
	entries := HeadlessEngineRegistry()
	modes := make([]string, 0, len(entries))
	for _, entry := range entries {
		modes = append(modes, entry.Mode)
	}
	return strings.Join(modes, "|")
}
