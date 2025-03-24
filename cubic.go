package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func doCubicNumbersSequence(config Config) {
	fmt.Println("Cubic Config IntOnly", config.IntOnly)
	fmt.Println("Cubic Config NoSmall", config.NoSmall)

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

		go workerCubic(heightChan, &wg, config)
	}

	// Enqueue heights for processing
	for height := 1; height <= maxHeight; height++ {
		heightChan <- height
	}

	// Close the channel to signal no more work
	close(heightChan)

	// Wait for all workers to complete
	wg.Wait()
	fmt.Printf("Total cubic execution time: %v\n", time.Since(start))
}

func workerCubic(heightChan <-chan int, wg *sync.WaitGroup, config Config) {
	defer wg.Done()

	// Process heights from the channel until it's closed
	for height := range heightChan {
		total := height * (2*height + 1) * (2*height + 1) * (2 * height) //nolint:mnd
		desc := "8h^4"

		if config.IntOnly {
			desc = "8h^3"
			total = (2*height + 1) * (2*height + 1) * (2*height + 1)
		}

		count, badCount, imaginaryCount, smallCount, doubleCount, notIrreducibleCount, threeRootsCount, twoRootsCount := processHeightCubic(height, config)
		p := message.NewPrinter(language.English)

		p.Printf("Processed "+desc+" %d, degree: 3, height: %d, count: %d, badCount: %d smallCount %d doubleCount %d imaginary Count %d twoRootsCount %d threeRootsCount %d !irreducibleCount %d percent %3.2f 1/log(h) %3.2f 1/sqrt(log(h)) %3.2f\n",
			total,
			height,
			count,
			badCount,
			smallCount,
			doubleCount,
			imaginaryCount,
			twoRootsCount,
			threeRootsCount,
			notIrreducibleCount,
			100.0*float64(badCount)/float64(count),
			100.0*1.0/float64(math.Log(float64(height))),
			100.0*1.0/float64(math.Sqrt(math.Log(float64(height)))))
	}
}

func processHeightCubic(height int, config Config) (count, badCount, imaginaryCount, smallCount, doubleCount, notIrreducibleCount, threeRootsCount, twoRootsCount int) { //nolint:gocognit,cyclop
	aLimitHigh := 1
	if !config.IntOnly {
		aLimitHigh = height
	}

	for a := 1; a <= aLimitHigh; a++ { // starting with 1 as - P and P have same root.
		for b := -height; b <= height; b++ {
			for c := -height; c <= height; c++ {
				for d := -height; c <= height; c++ {
					if d == 0 {
						notIrreducibleCount++

						continue
					}

					roots, irreducible := getCubicRoots(a, b, c, d)
					if !irreducible {
						notIrreducibleCount++

						continue
					}

					switch {
					case len(roots) == 0:
						panic("no roots for a cubic?")
					case len(roots) == 1:
						imaginaryCount++
					case len(roots) == 2: // nolint:mnd
						twoRootsCount++
					default:
						threeRootsCount++
					}

					for _, root := range roots {
						if rand.Intn(1000) == 0 { // nolint:gosec,mnd
							testRootsCubic(a, b, c, d, root)
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
	}

	return //nolint:nakedret
}

// It also indicates whether the polynomial is irreducible over the reals
// Parameters:
//
//	a, b, c, d: coefficients of the cubic equation
//
// Returns:
//
//	[]float64: slice containing all real roots (0 to 3 values)
//	bool: true if the polynomial is irreducible over the reals
func getCubicRoots(a, b, c, d int) ([]float64, bool) { //nolint:cyclop
	// Handle the case where a is 0 (quadratic equation)
	if a == 0 {
		roots := getQuadraticRoots([]int{b, c, d})
		// For cubics, if we have fewer than 3 real roots,
		// then the polynomial is irreducible over the reals
		isIrreducible := len(roots) < 3 //nolint:mnd

		return roots, isIrreducible
	}

	// Convert to depressed cubic form t³ + pt + q = 0
	// by substituting x = t - b/(3a)
	fa := float64(a)
	fb := float64(b)
	fc := float64(c)
	fd := float64(d)

	// Calculate intermediate values
	b2 := fb * fb
	ac := fa * fc
	p := (3.0*ac - b2) / (3.0 * fa * fa)
	q := (2.0*b2*fb - 9.0*fa*fb*fc + 27.0*fa*fa*fd) / (27.0 * fa * fa * fa)

	// Handle different cases based on discriminant
	discriminant := (q*q/4.0 + p*p*p/27.0)

	roots := []float64{}

	// Special case: p = q = 0, triple root
	if math.Abs(p) < 1e-10 && math.Abs(q) < 1e-10 {
		roots = append(roots, -fb/(3.0*fa)) //nolint:mnd

		return roots, true
	}

	// Calculate the shift for converting back from depressed form
	shift := fb / (3.0 * fa) //nolint:mnd

	if discriminant < 0 {
		// Three real roots case (casus irreducibilis)
		// Use trigonometric solution
		sqrtNegP := math.Sqrt(-p)
		phi := math.Acos(-q / (2.0 * sqrtNegP * sqrtNegP * sqrtNegP))

		// All three roots
		for k := range 3 {
			root := 2.0*sqrtNegP*math.Cos((phi+2.0*math.Pi*float64(k))/3.0) - shift //nolint:mnd
			roots = append(roots, root)
		}
	} else {
		// One real root (discriminant > 0) or at least one repeated root (discriminant = 0)
		u := cbrt(-q/2.0 + math.Sqrt(discriminant))
		v := cbrt(-q/2.0 - math.Sqrt(discriminant))

		// First root is always real
		roots = append(roots, u+v-shift)

		// Check if we have repeated roots (discriminant = 0)
		if math.Abs(discriminant) < 1e-10 { //nolint:mnd
			// One single and one double root
			repeatedRoot := -(u+v)/2.0 - shift //nolint:mnd
			roots = append(roots, repeatedRoot)

			// Add repeated root only once more if it's truly a double root
			if math.Abs(p) > 1e-10 { //nolint:mnd
				roots = append(roots, repeatedRoot)
			}
		}
	}

	// Sort the roots for consistency
	sort.Float64s(roots)

	// Remove duplicates within floating-point precision
	if len(roots) > 1 {
		uniqueRoots := []float64{roots[0]}

		for i := 1; i < len(roots); i++ {
			if math.Abs(roots[i]-roots[i-1]) > 1e-10 { //nolint:mnd
				uniqueRoots = append(uniqueRoots, roots[i])
			}
		}

		roots = uniqueRoots
	}

	return roots, len(roots) < 3 //nolint:mnd
}

// Helper function for cubic root that handles negative numbers.
func cbrt(x float64) float64 {
	if x < 0 {
		return -math.Pow(-x, 1.0/3.0) //nolint:mnd
	}

	return math.Pow(x, 1.0/3.0) //nolint:mnd
}
