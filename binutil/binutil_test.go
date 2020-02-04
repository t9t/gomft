package binutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/t9t/gomft/binutil"
)

func TestIsOnlyZeroesYes(t *testing.T) {
	assert.True(t, binutil.IsOnlyZeroes([]byte{0, 0, 0, 0, 0, 0}))
}

func TestIsOnlyZeroesNo(t *testing.T) {
	assert.False(t, binutil.IsOnlyZeroes([]byte{0, 0, 0, 0, 0, 1}))
}
