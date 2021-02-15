package kana

import (
	"math/rand"
	"sort"
)

// CreateWeights creates weights for a given set of kana.
func CreateWeights(values map[string]string) map[string]float64 {
	output := make(map[string]float64)
	for key := range values {
		output[key] = WeightDefault
	}
	return output
}

// IncreaseWeight increases the weight for a given value.
func IncreaseWeight(weights map[string]float64, key string) {
	if weight, ok := weights[key]; ok {
		if weight < WeightMax {
			weights[key] = weight * WeightIncreaseFactor
		}
	}
}

// DecreaseWeight decreases the weight for a given value.
func DecreaseWeight(weights map[string]float64, key string) {
	if weight, ok := weights[key]; ok {
		if weight <= WeightMin {
			return
		}
		weights[key] = weight / WeightDecreaseFactor
	}
}

// SelectCount returns the first N elements from a given combined values set.
func SelectCount(values map[string]string, count int) map[string]string {
	if count == 0 || len(values) <= count {
		return values
	}
	output := make(map[string]string)
	for key, value := range values {
		output[key] = value
		if len(output) == count {
			break
		}
	}
	return output
}

// SelectWeighted selects a kana and a roman from a given set of values and weights.
func SelectWeighted(values map[string]string, weights map[string]float64) (kana, roman string) {
	// collect "weighted" choices
	type weightedChoice struct {
		Key    string
		Weight float64
	}
	var keys []weightedChoice
	for key := range values {
		keys = append(keys, weightedChoice{
			Key:    key,
			Weight: weights[key],
		})
	}

	// sort by weight ascending
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Weight < keys[j].Weight
	})

	// sum all the weights, assign to indexes
	totals := make([]float64, len(keys))
	var runningTotal float64
	for index, wc := range keys {
		runningTotal += wc.Weight
		totals[index] = runningTotal
	}
	randomValue := rand.Float64() * runningTotal
	randomIndex := sort.SearchFloat64s(totals, randomValue)

	kana = keys[randomIndex].Key
	roman = values[kana]
	return
}

// Merge merges variadic sets of values.
func Merge(sets ...map[string]string) map[string]string {
	output := make(map[string]string)
	for _, set := range sets {
		for key, value := range set {
			output[key] = value
		}
	}
	return output
}

// ListHas returns if a value is present in a list
func ListHas(list []string, value string) bool {
	for _, listValue := range list {
		if listValue == value {
			return true
		}
	}
	return false
}

// ListAddFixedLength adds a value to a given list
func ListAddFixedLength(list []string, value string, max int) []string {
	list = append(list, value)
	if len(list) < max {
		return list
	}
	return list[1:]
}
