package main

import (
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

// Config represents the configuration parameters
type Config struct {
	MaxHeight  int
	NumWorkers int
	Profile    bool   // Whether to enable profiling
	ProfileOut string // Path to save the profile
	IntOnly    bool   // algebraic ints or algebraic numbers
	NoSmall    bool   // only for |q| > Height/2
}

// Result holds the statistics for a specific height
type Result struct {
	Height   int
	Count    int
	BadCount int
}

// Rest of your code remains the same...

func main() {

	config := Config{
		MaxHeight:  500,
		NumWorkers: 8,
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

	config.IntOnly = false
	config.NoSmall = false
	doAlgebraicNumbersSequence(config)

	config.IntOnly = true
	config.NoSmall = false
	doAlgebraicNumbersSequence(config)

	config.IntOnly = false
	config.NoSmall = true
	doAlgebraicNumbersSequence(config)

	config.IntOnly = true
	config.NoSmall = true
	doAlgebraicNumbersSequence(config)

}

func doAlgebraicNumbersSequence(config Config) {

	fmt.Println("Config IntOnly", config.IntOnly)
	fmt.Println("Config NoSmall", config.NoSmall)

	start := time.Now()

	maxHeight := config.MaxHeight
	numWorkers := config.NumWorkers

	// Create a channel to distribute work to workers
	heightChan := make(chan int, maxHeight)

	// Create a wait group to wait for all workers to complete
	var wg sync.WaitGroup

	// Launch worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(heightChan, &wg, config)
	}

	// Enqueue heights for processing
	for height := 1; height <= maxHeight; height++ {
		heightChan <- height
	}

	// Close the channel to signal no more work
	close(heightChan)

	// Wait for all workers to complete
	wg.Wait()
	fmt.Printf("Total execution time: %v\n", time.Since(start))
}

func worker(heightChan <-chan int, wg *sync.WaitGroup, config Config) {
	defer wg.Done()

	// Process heights from the channel until it's closed
	for height := range heightChan {
		count, badCount := processHeight(height, config)
		fmt.Printf("Processed height: %d, count: %d, badCount: %d percent %3.2f 1/log(h) %3.2f\n",
			height,
			count,
			badCount,
			100.0*float64(badCount)/float64(count),
			100.0*1.0/float64(math.Log(float64(height))))
	}
}

func processHeight(height int, config Config) (count, badCount int) {
	aLimitLow := 1
	aLimitHigh := 1
	if !config.IntOnly {
		aLimitLow = -height
		aLimitHigh = height
	}
	// a := 1
	for a := aLimitLow; a <= aLimitHigh; a++ {
		for b := -height; b <= height; b++ {
			for c := -height; c <= height; c++ {
				coeffs := []int{a, b, c}
				if isIrreducible(coeffs) {
					roots := getRealRoots(coeffs)
					for _, root := range roots {

						if config.NoSmall && math.Abs(root) < float64(height)/2.0 {
							continue
						}
						count++
						val := math.Mod(root, 1)
						if val < 0 {
							val += 1 // Ensure val is in [0,1)
						}
						if val < 1/(math.Log(math.Abs(root))) {
							// fmt.Println("Height", height, "root", root, "val", val, "log", 1/(math.Log(math.Abs(root))))
							badCount++
						}
					}
				}
			}
		}
	}
	return count, badCount
}

func getPolynomialHeight(coeffs []int) int {
	maxCoeff := 0
	for _, c := range coeffs {
		if absC := int(math.Abs(float64(c))); absC > maxCoeff {
			maxCoeff = absC
		}
	}
	return maxCoeff
}

func isIrreducible(coeffs []int) bool {
	a, b, c := coeffs[0], coeffs[1], coeffs[2]
	discriminant := b*b - 4*a*c

	if discriminant < 0 {
		return true
	}

	sqrtDisc := int(math.Sqrt(float64(discriminant)))
	return sqrtDisc*sqrtDisc != discriminant
}

func getRealRoots(coeffs []int) []float64 {
	a, b, c := float64(coeffs[0]), float64(coeffs[1]), float64(coeffs[2])
	discriminant := b*b - 4*a*c

	if discriminant < 0 {
		return []float64{}
	}

	sqrtDisc := math.Sqrt(discriminant)
	root1 := (-b + sqrtDisc) / (2 * a)
	root2 := (-b - sqrtDisc) / (2 * a)

	// Sort the roots
	if root1 > root2 {
		return []float64{root2, root1}
	}
	return []float64{root1, root2}
}
