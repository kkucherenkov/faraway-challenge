package transport

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseMessage(t *testing.T) {
	t.Run("Empty message can't be parsed", func(t *testing.T) {
		input := ""
		msg, err := ParseMessage(input)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "can't recognize the type of message", err.Error())
	})

	t.Run("Invalid message format can't be parsed", func(t *testing.T) {
		input := fmt.Sprintf("%s%s", separator, separator)
		msg, err := ParseMessage(input)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "message malformed", err.Error())
	})

	t.Run("Wrong message type", func(t *testing.T) {
		input := fmt.Sprintf("%s%s%s", "type", separator, "data")
		msg, err := ParseMessage(input)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "can't recognize the type of message", err.Error())
	})

	t.Run("Challenge request message: parsed", func(t *testing.T) {
		input := fmt.Sprintf("%d%s", GetChallenge, separator)
		msg, err := ParseMessage(input)
		require.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, 1, msg.Type)
		assert.Empty(t, msg.Data)
	})

	t.Run("Correct message: parsed", func(t *testing.T) {
		input := fmt.Sprintf("%d%s%s", GetChallenge, separator, "data")
		msg, err := ParseMessage(input)
		require.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, 1, msg.Type)
		assert.Equal(t, "data", msg.Data)
	})

}
