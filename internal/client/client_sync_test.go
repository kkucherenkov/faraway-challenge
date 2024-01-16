package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kkucherenkov/faraway-challenge/internal/pkg/config"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/pow"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleConnection(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	t.Parallel()

	ctx := context.Background()
	cfg := config.Config{Pow: config.Pow{Complexity: 2, Expiration: 30000, MaxIterations: 100000}, Service: config.Service{Port: "8080", Host: "localhost", DataFile: ""}, Cache: config.Cache{Port: "1234", Host: "localhost"}}
	ctx = context.WithValue(ctx, "config", &cfg)
	ctx = context.WithValue(ctx, "logger", logger)

	t.Run("Write error", func(t *testing.T) {
		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				return 0, fmt.Errorf("test write error")
			},
		}
		_, err := HandleConnection(ctx, mock, mock)
		assert.Error(t, err)
		assert.Equal(t, "err send request: test write error", err.Error())
	})

	t.Run("Read error", func(t *testing.T) {
		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				return 0, nil
			},
			readFunc: func(p []byte) (int, error) {
				return 0, fmt.Errorf("test read error")
			},
		}
		_, err := HandleConnection(ctx, mock, mock)
		assert.Error(t, err)
		assert.Equal(t, "err read msg: test read error", err.Error())
	})

	t.Run("Read response in bad format", func(t *testing.T) {
		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				return 0, nil
			},
			readFunc: func(p []byte) (int, error) {
				return fillTestReadBytes("||\n", p), nil
			},
		}
		_, err := HandleConnection(ctx, mock, mock)
		assert.Error(t, err)
		assert.Equal(t, "err parse msg: message malformed", err.Error())
	})

	t.Run("Read response with hashcash in bad format", func(t *testing.T) {
		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				return 0, nil
			},
			readFunc: func(p []byte) (int, error) {
				return fillTestReadBytes(fmt.Sprintf("%d|{wrong_json}\n", transport.Challenge), p), nil
			},
		}
		_, err := HandleConnection(ctx, mock, mock)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "err parse hashcash"))
	})

	t.Run("Success", func(t *testing.T) {
		date := time.Date(2022, 3, 13, 2, 30, 0, 0, time.UTC)
		hashcash := pow.HashCashData{
			Version:    1,
			ZerosCount: 3,
			Date:       date.Unix(),
			Resource:   "client1",
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123460))),
			Counter:    0,
		}

		// counter for reading attempts to change content
		readAttempt := 0

		writeAttempt := 0

		mock := mockConnection{
			writeFunc: func(p []byte) (int, error) {
				if writeAttempt == 0 {
					writeAttempt++
					assert.Equal(t, "1|\n", string(p))
				} else {
					msg, err := transport.ParseMessage(string(p))
					require.NoError(t, err)
					var writtenHashcash pow.HashCashData
					err = json.Unmarshal([]byte(msg.Data), &writtenHashcash)
					require.NoError(t, err)
					// checking that counter increased
					assert.Equal(t, 5001, writtenHashcash.Counter)
					_, err = writtenHashcash.Compute(0)
					assert.NoError(t, err)
				}
				return 0, nil
			},
			readFunc: func(p []byte) (int, error) {
				if readAttempt == 0 {
					marshaled, err := json.Marshal(hashcash)
					require.NoError(t, err)
					readAttempt++
					return fillTestReadBytes(fmt.Sprintf("%d|%s\n", transport.Challenge, string(marshaled)), p), nil
				} else {
					// second read, send quote
					return fillTestReadBytes(fmt.Sprintf("%d|test quote\n", transport.Challenge), p), nil
				}
			},
		}
		response, err := HandleConnection(ctx, mock, mock)
		assert.NoError(t, err)
		assert.Equal(t, "test quote", response)
	})
}
