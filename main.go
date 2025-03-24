package main

import (
	"fmt"
	"os"
	"runtime/pprof"
)

// Config represents the configuration parameters.
type Config struct {
	MaxHeight  int
	NumWorkers int
	Profile    bool   // Whether to enable profiling
	ProfileOut string // Path to save the profile
	IntOnly    bool   // algebraic ints or algebraic numbers
	NoSmall    bool   // only for |q| > Height/2
}

// Result holds the statistics for a specific height.
type Result struct {
	Height   int
	Count    int
	BadCount int
}

// Rest of your code remains the same...

func main() {
	config := Config{
		MaxHeight:  500,              //nolint:mnd
		NumWorkers: 12,               //nolint:mnd
		Profile:    true,             // Enable profiling
		ProfileOut: "algstats.pprof", // Profile output file
		IntOnly:    true,
		NoSmall:    true,
	}

	// Start CPU profiling if enabled
	if config.Profile {
		f, err := os.Create(config.ProfileOut)
		if err != nil {
			fmt.Printf("Error creating profile file: %v\n", err)

			return
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Printf("Error starting CPU profile: %v\n", err)

			return
		}
		defer pprof.StopCPUProfile()
		fmt.Println("CPU profiling enabled, writing to", config.ProfileOut)
	}

	config.IntOnly = true
	config.NoSmall = true
	doQuadraticNumbersSequence(config)
	doCubicNumbersSequence(config)

	config.IntOnly = false
	config.NoSmall = true
	doQuadraticNumbersSequence(config)
	doCubicNumbersSequence(config)
}
