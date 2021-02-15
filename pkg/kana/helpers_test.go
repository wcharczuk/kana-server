package kana_test

import (
	"fmt"
	"testing"

	"github.com/blend/go-sdk/assert"

	"github.com/wcharczuk/kana-server/pkg/kana"
)

func Test_SelectWeighted_uniform(t *testing.T) {
	its := assert.New(t)

	kvs := GenerateKeyValues(10)
	kvws := GenerateKeyWeightsDefault(10)

	var key, value string
	selections := make(map[string]int)
	for x := 0; x < 1024; x++ {
		key, value = kana.SelectWeighted(kvs, kvws)
		its.Equal(value, kvs[key])
		selections[key]++
	}

	fudge := 25
	expectedMin := (1024 / 10) - fudge
	expectedMax := (1024 / 10) + fudge
	var total int
	for _, count := range selections {
		total += count
		its.True(count < expectedMax, fmt.Sprintf("%d should be < %v and > %v", count, expectedMax, expectedMin))
		its.True(count > expectedMin, fmt.Sprintf("%d should be < %v and > %v", count, expectedMax, expectedMin))
	}
}

func Test_SelectWeighted_increasing(t *testing.T) {
	its := assert.New(t)

	kvs := GenerateKeyValues(10)
	kvws := GenerateKeyWeightsIncreasing(10)

	var key, value string
	selections := make(map[string]int)
	for x := 0; x < 1024; x++ {
		key, value = kana.SelectWeighted(kvs, kvws)
		its.Equal(value, kvs[key])
		selections[key]++
	}

	var previous int
	for x := 0; x < 10; x++ {
		key := "k" + fmt.Sprint(x)
		count := selections[key]
		its.True(count >= previous, fmt.Sprintf("%d should be > than %d, %#v", count, previous, selections))
	}
}

func Test_SelectWeighted_decreasing(t *testing.T) {
	its := assert.New(t)

	kvs := GenerateKeyValues(10)
	kvws := GenerateKeyWeightsIncreasing(10)

	var key, value string
	selections := make(map[string]int)
	for x := 0; x < 1024; x++ {
		key, value = kana.SelectWeighted(kvs, kvws)
		its.Equal(value, kvs[key])
		selections[key]++
	}

	var previous int
	for x := 0; x < 10; x++ {
		key := "k" + fmt.Sprint(x)
		count := selections[key]
		its.True(count >= previous, fmt.Sprintf("%d should be > than %d, %#v", count, previous, selections))
	}
}
