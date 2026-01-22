// pkg/gui/tabs/state.go
package tabs

import "demo6-eda/pkg/compiler"

// AppState holds shared state across tabs
type AppState struct {
	CurrentMapping *compiler.CrossbarMapping
	WeightsLoaded  bool
	Compiled       bool
}
