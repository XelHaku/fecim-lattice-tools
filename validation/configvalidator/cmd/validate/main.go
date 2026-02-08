// Package main provides a CLI tool for validating FeCIM configuration files.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fecim-lattice-tools/validation/configvalidator"
)

func main() {
	var (
		recursive   = flag.Bool("r", false, "Recursively validate all JSON files in directories")
		showWarnings = flag.Bool("w", false, "Show warnings (not just errors)")
		summary     = flag.Bool("s", false, "Show summary only (no individual file results)")
		quiet       = flag.Bool("q", false, "Quiet mode (only exit code)")
	)
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file.json|directory> ...\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Validates FeCIM configuration JSON files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSupported config types:\n")
		fmt.Fprintf(os.Stderr, "  - calibration:    Ferroelectric calibration data\n")
		fmt.Fprintf(os.Stderr, "  - preisach_state: Preisach hysteron states\n")
		fmt.Fprintf(os.Stderr, "  - array_design:   Crossbar array designs\n")
		fmt.Fprintf(os.Stderr, "  - weight_matrix:  Neural network weight matrices\n")
		fmt.Fprintf(os.Stderr, "  - openlane:       OpenLane ASIC flow configs\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s data/calibrations/fecim_hzo.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -r data/\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -w -s .\n", os.Args[0])
	}
	
	flag.Parse()
	
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	
	var allResults []*configvalidator.ValidationResult
	
	// Process each argument
	for _, arg := range flag.Args() {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot access %s: %v\n", arg, err)
			continue
		}
		
		if info.IsDir() {
			if *recursive {
				results, err := configvalidator.ValidateDirectory(arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error validating directory %s: %v\n", arg, err)
					continue
				}
				allResults = append(allResults, results...)
			} else {
				// Just validate JSON files in the immediate directory
				entries, err := os.ReadDir(arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", arg, err)
					continue
				}
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
						path := filepath.Join(arg, entry.Name())
						result, err := configvalidator.ValidateFile(path)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error validating %s: %v\n", path, err)
							continue
						}
						allResults = append(allResults, result)
					}
				}
			}
		} else {
			result, err := configvalidator.ValidateFile(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error validating %s: %v\n", arg, err)
				continue
			}
			allResults = append(allResults, result)
		}
	}
	
	// Process results
	var totalFiles, validFiles, invalidFiles int
	var totalErrors, totalWarnings int
	
	for _, result := range allResults {
		totalFiles++
		if result.Valid {
			validFiles++
		} else {
			invalidFiles++
		}
		totalErrors += len(result.Errors)
		totalWarnings += len(result.Warnings)
		
		// Print individual results unless in quiet or summary mode
		if !*quiet && !*summary {
			if !result.Valid || (*showWarnings && len(result.Warnings) > 0) {
				fmt.Println(result.String())
				fmt.Println(strings.Repeat("-", 60))
			}
		}
	}
	
	// Print summary
	if !*quiet {
		if *summary || totalFiles > 1 {
			fmt.Printf("\n=== Validation Summary ===\n")
			fmt.Printf("Total files:  %d\n", totalFiles)
			fmt.Printf("Valid:        %d\n", validFiles)
			fmt.Printf("Invalid:      %d\n", invalidFiles)
			fmt.Printf("Total errors: %d\n", totalErrors)
			if *showWarnings {
				fmt.Printf("Total warnings: %d\n", totalWarnings)
			}
		}
	}
	
	// Exit with appropriate code
	if invalidFiles > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
