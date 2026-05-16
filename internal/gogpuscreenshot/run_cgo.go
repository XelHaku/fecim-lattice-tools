//go:build cgo

package gogpuscreenshot

import "fmt"

func Run(args []string) error {
	return fmt.Errorf("gogpu screenshotter requires CGO_ENABLED=0; rebuild with CGO_ENABLED=0 go run ./cmd/fecim-screenshotter")
}
