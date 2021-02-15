package kana_test

import (
	"fmt"

	"github.com/wcharczuk/kana-server/pkg/kana"
)

// GenerateKeyValues generates N key value pairs.
func GenerateKeyValues(n int) map[string]string {
	output := make(map[string]string)
	var xs string
	for x := 0; x < n; x++ {
		xs = fmt.Sprint(x)
		output["k"+xs] = "v" + xs
	}
	return output
}

// GenerateKeyWeightsDefault generates N uniform key weights.
func GenerateKeyWeightsDefault(n int) map[string]float64 {
	output := make(map[string]float64)
	var xs string
	for x := 0; x < n; x++ {
		xs = fmt.Sprint(x)
		output["k"+xs] = kana.WeightDefault
	}
	return output
}

// GenerateKeyWeightsIncreasing generates N uniform key weights.
func GenerateKeyWeightsIncreasing(n int) map[string]float64 {
	output := make(map[string]float64)
	var xs string
	for x := 0; x < n; x++ {
		xs = fmt.Sprint(x)
		output["k"+xs] = kana.WeightDefault
	}
	for x := 0; x < n; x++ {
		xs = fmt.Sprint(x)
		for y := 0; y < x; y++ {
			kana.IncreaseWeight(output, "k"+xs)
		}
	}
	return output
}
