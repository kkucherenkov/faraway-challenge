package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/pow"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestGetChallenge(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	t.Parallel()

	t.Run("Get challenge write error", func(t *testing.T) {
		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				return 0, fmt.Errorf("test write error")
			},
		}
		_, err := GetChallenge(mock, mock, logger)
		assert.Error(t, err)
		assert.Equal(t, "err send request: test write error", err.Error())
	})

	t.Run("Get quote write error", func(t *testing.T) {
		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				return 0, fmt.Errorf("test write error")
			},
		}
		_, err := GetQuote(mock, mock, logger, "")
		assert.Error(t, err)
		assert.Equal(t, "err send request: test write error", err.Error())
	})

	t.Run("Get challenge read error", func(t *testing.T) {
		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				return 0, nil
			},
			readFunc: func(p []byte) (int, error) {
				return 0, fmt.Errorf("test read error")
			},
		}
		_, err := GetChallenge(mock, mock, logger)
		assert.Error(t, err)
		assert.Equal(t, "err read msg: test read error", err.Error())
	})

	t.Run("Get quote read error", func(t *testing.T) {
		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				return 0, nil
			},
			readFunc: func(p []byte) (int, error) {
				return 0, fmt.Errorf("test read error")
			},
		}
		_, err := GetQuote(mock, mock, logger, "")
		assert.Error(t, err)
		assert.Equal(t, "err read msg: test read error", err.Error())
	})

	t.Run("Get challenge success", func(t *testing.T) {
		date := time.Date(2024, 1, 14, 0, 0, 10, 0, time.UTC)
		hashcash := pow.HashCashData{
			Version:    1,
			ZerosCount: 3,
			Date:       date.Unix(),
			Resource:   "client1",
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123460))),
			Counter:    0,
		}
		writeAttempt := 0

		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				if writeAttempt == 0 {
					assert.Equal(t, "1|\n", string(p))
				}
				if writeAttempt == 1 {
					assert.Equal(t, "0|\n", string(p))
				}
				writeAttempt++
				return 0, nil
			},
			readFunc: func(p []byte) (int, error) {
				marshaled, err := json.Marshal(hashcash)
				require.NoError(t, err)
				return fillTestReadBytes(fmt.Sprintf("%d|%s\n", transport.Challenge, string(marshaled)), p), nil
			},
		}
		response, err := GetChallenge(mock, mock, logger)
		assert.NoError(t, err)
		assert.Equal(t, "client1", response.Resource)
	})

	t.Run("Get quote success", func(t *testing.T) {
		writeAttempt := 0
		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				if writeAttempt == 0 {
					assert.Equal(t, "3|solution\n", string(p))
				}
				if writeAttempt == 1 {
					assert.Equal(t, "0|\n", string(p))
				}
				writeAttempt++
				return 0, nil
			},
			readFunc: func(p []byte) (int, error) {
				return fillTestReadBytes(fmt.Sprintf("%d|test quote\n", transport.Challenge), p), nil
			},
		}
		response, err := GetQuote(mock, mock, logger, "solution")
		assert.NoError(t, err)
		assert.Equal(t, "test quote", response)
	})
}
