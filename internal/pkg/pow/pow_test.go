package pow

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVerifySolution(t *testing.T) {
	t.Parallel()

	t.Run("Correct solution", func(t *testing.T) {
		result := Verify("0000abcdef", 4)
		assert.True(t, result)
	})

	t.Run("Too short hash", func(t *testing.T) {
		result := Verify("0000", 5)
		assert.False(t, result)
	})

	t.Run("Incorrect solution", func(t *testing.T) {
		result := Verify("000abcdef", 4)
		assert.False(t, result)
	})
}
