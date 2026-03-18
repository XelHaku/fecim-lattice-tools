package main

// globalSeed holds the --seed value for deterministic simulation replay.
// When zero (default), solvers use non-deterministic seeds.
var globalSeed int64

// GlobalSeed returns the seed set via --seed flag. Zero means non-deterministic.
func GlobalSeed() int64 {
	return globalSeed
}
