package main

import (
	"fmt"
	"math"
)

func testRootsCubic(
	a int,
	b int,
	c int,
	d int,
	root float64,
) {
	tol := math.MaxFloat64
	for _, coef := range []int{a, b, c, d} {
		if coef != 0 && math.Abs(float64(coef)) < tol {
			tol = math.Abs(float64(coef))
		}
	}

	tol *= 0.002 // 0.2% tolerance

	// Evaluate polynomial at the root
	fa, fb, fc, fd := float64(a), float64(b), float64(c), float64(d)
	result := fa*math.Pow(root, 3) + fb*math.Pow(root, 2) + fc*root + fd // nolint:mnd

	// If any root fails, return false
	if math.Abs(result) > tol {
		fmt.Printf("root not a root %d %d %d %d %3.52f\n", a, b, c, d, root)
	}
}
