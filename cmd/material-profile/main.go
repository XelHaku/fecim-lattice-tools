// Command material-profile exports material physics parameters as JSON.
//
// Usage: material-profile -material fecim_hzo -output profile.json
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"fecim-lattice-tools/shared/physics"
)

func main() {
	profile := flag.String("profile", "pr", "material profile: pr|nightly")
	mode := flag.String("mode", "list", "mode: list|version")
	sep := flag.String("sep", "\n", "separator for list output")
	flag.Parse()

	switch *mode {
	case "version":
		fmt.Println(physics.MaterialProfileVersion)
		return
	case "list":
		mats, err := physics.RequiredMaterialsForProfile(physics.MaterialProfileName(*profile))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Print(strings.Join(mats, *sep))
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(2)
	}
}
