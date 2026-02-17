// Command material-profile exports material physics parameters as JSON.
//
// Usage: material-profile -material fecim_hzo -output profile.json
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"fecim-lattice-tools/shared/physics"
)

func run(out, errOut io.Writer, profile, mode, sep string) int {
	switch mode {
	case "version":
		fmt.Fprintln(out, physics.MaterialProfileVersion)
		return 0
	case "list":
		mats, err := physics.RequiredMaterialsForProfile(physics.MaterialProfileName(profile))
		if err != nil {
			fmt.Fprintln(errOut, err)
			return 2
		}
		fmt.Fprint(out, strings.Join(mats, sep))
		return 0
	default:
		fmt.Fprintf(errOut, "unknown mode %q\n", mode)
		return 2
	}
}

func main() {
	profile := flag.String("profile", "pr", "material profile: pr|nightly")
	mode := flag.String("mode", "list", "mode: list|version")
	sep := flag.String("sep", "\n", "separator for list output")
	flag.Parse()

	os.Exit(run(os.Stdout, os.Stderr, *profile, *mode, *sep))
}
