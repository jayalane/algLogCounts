package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func doQuadraticNumbersSequence(config Config) {
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
	for range numWorkers {
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
	fmt.Printf("Total degree 2 execution time: %v\n", time.Since(start))
}

func worker(heightChan <-chan int, wg *sync.WaitGroup, config Config) {
	defer wg.Done()

	// Process heights from the channel until it's closed
	for height := range heightChan {
		total := height * (2*height + 1) * (2*height + 1) // nolint:mnd
		desc := "4h^3"

		if config.IntOnly {
			total = (2*height + 1) * (2 * height) // nolint:mnd
			desc = "4h^2"
		}

		count, badCount, imaginaryCount, smallCount, doubleCount, notIrreducibleCount, twoRootsCount := processHeight(height, config)
		p := message.NewPrinter(language.English)

		p.Printf("Processed "+desc+" %d, degree: 2, height: %d, count: %d, badCount: %d smallCount %d doubleCount %d imaginary Count %d twoRootsCount %d !irreducibleCount %d percent %3.2f 1/log(h) %3.2f 1/sqrt(log(h)) %3.2f\n",
			total,
			height,
			count,
			badCount,
			smallCount,
			doubleCount,
			imaginaryCount,
			twoRootsCount,
			notIrreducibleCount,
			100.0*float64(badCount)/float64(count),
			100.0*1.0/float64(math.Log(float64(height))),
			100.0*1.0/float64(math.Sqrt(math.Log(float64(height)))))
	}
}

func processHeight(height int, config Config) (count, badCount, imaginaryCount, smallCount, doubleCount, notIrreducibleCount, twoRootsCount int) { //nolint:gocognit,cyclop
	aLimitHigh := 1
	if !config.IntOnly {
		aLimitHigh = height
	}

	for a := 1; a <= aLimitHigh; a++ { // starting with 1 as - P and P have same root.
		for b := -height; b <= height; b++ {
			for c := -height; c <= height; c++ {
				coeffs := []int{a, b, c}

				if c == 0 {
					notIrreducibleCount++

					continue
				}

				if !isIrreducible(coeffs) {
					notIrreducibleCount++

					continue
				}

				roots := getQuadraticRoots(coeffs)

				switch {
				case len(roots) == 0:
					imaginaryCount++
				case len(roots) == 1:
					doubleCount++
				default:
					twoRootsCount++
				}

				for _, root := range roots {
					if rand.Intn(1000) == 0 { // nolint:gosec,mnd
						testRootsCubic(0, a, b, c, root)
					}

					if config.NoSmall && math.Abs(root) < float64(math.Abs(float64(height)))/2.0 {
						smallCount++

						continue
					}

					count++

					val := math.Mod(root, 1)
					if val < 0 {
						val++ // Ensure val is in [0,1)
					}

					if 1/val > math.Abs(math.Log(math.Abs(root))) {
						// fmt.Println("Height", height, "root", root, "val", val, "1/val", 1.0/val, "log", math.Abs(math.Log(math.Abs(root))))
						badCount++
					}
				}
			}
		}
	}

	return //nolint:nakedret
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

func getQuadraticRoots(coeffs []int) []float64 {
	a, b, c := float64(coeffs[0]), float64(coeffs[1]), float64(coeffs[2])

	discriminantInt := coeffs[1]*coeffs[1] - 4*coeffs[0]*coeffs[2]
	if discriminantInt == 0 {
		root := (-b) / (2 * a) //nolint:mnd

		return []float64{root}
	}

	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return []float64{}
	}

	sqrtDisc := math.Sqrt(discriminant)
	root1 := (-b + sqrtDisc) / (2 * a) //nolint:mnd
	root2 := (-b - sqrtDisc) / (2 * a) //nolint:mnd

	// Sort the roots
	if root1 > root2 {
		return []float64{root2, root1}
	}

	return []float64{root1, root2}
}
