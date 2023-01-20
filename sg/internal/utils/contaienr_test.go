package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Map(t *testing.T) {
	assert.Equal(
		t,
		[]int{},
		Map([]int{}, func(x int) int { return x }),
	)

	assert.Equal(
		t,
		[]int{1, 2, 3},
		Map([]int{1, 2, 3}, func(x int) int { return x }),
	)

	assert.Equal(
		t,
		[]int{2, 4, 6},
		Map([]int{1, 2, 3}, func(x int) int { return x * 2 }),
	)

	assert.Equal(
		t,
		[]string{"1", "2", "3"},
		Map([]int{1, 2, 3}, func(x int) string { return fmt.Sprint(x) }),
	)
}

func Test_Filter(t *testing.T) {
	assert.Equal(
		t,
		[]int{},
		Filter([]int{}, func(x int) bool { return true }),
	)

	assert.Equal(
		t,
		[]int{1, 2, 3},
		Filter([]int{1, 2, 3}, func(x int) bool { return true }),
	)

	assert.Equal(
		t,
		[]int{},
		Filter([]int{1, 2, 3}, func(x int) bool { return false }),
	)

	assert.Equal(
		t,
		[]int{1, 2},
		Filter([]int{1, 2, 3}, func(x int) bool { return x < 3 }),
	)
}
