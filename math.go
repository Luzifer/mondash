package main

import (
	"math"
	"sort"
)

func absoluteDeviation(values []float64) []float64 {
	medianValue := median(values)

	deviation := make([]float64, len(values))

	for i := range values {
		deviation[i] = math.Abs(values[i] - medianValue)
	}

	return deviation
}

func median(values []float64) float64 {
	sort.Float64s(values)

	if len(values) == 1 {
		return values[0]
	}

	// If even, take an average
	if len(values)%2 == 0 {
		return 0.5*values[len(values)/2] + 0.5*values[len(values)/2-1]
	}

	return values[len(values)/2-1]
}
