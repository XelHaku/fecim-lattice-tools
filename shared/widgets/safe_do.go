package widgets

// SafeDo executes fn in the Fyne runtime context (when available) and
// serializes concurrent UI updates so race-enabled tests do not run Fyne widget
// operations on multiple goroutines at once.
func SafeDo(fn func()) {
	safeUIUpdate(fn)
}
