package service

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

	"github.com/kkucherenkov/faraway-challenge/internal/pkg/cache"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/config"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/pow"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/storage"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClock - mock for Clock interface (to work with predefined Now)
type MockClock struct {
	NowFunc func() time.Time
}

func (m *MockClock) Now() time.Time {
	if m.NowFunc != nil {
		return m.NowFunc()
	}
	return time.Now()
}

func TestProcessRequest(t *testing.T) {
	const randKey = "123460"
	const clientInfo = "test client"

	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Config{Pow: config.Pow{Complexity: 2, Expiration: 30000, MaxIterations: 100}, Service: config.Service{Port: "8080", Host: "localhost", DataFile: ""}, Cache: config.Cache{Port: "1234", Host: "localhost"}}
	mockClock := MockClock{}
	requestCache := cache.NewLocalCache(&mockClock)
	quotes := storage.CreateStorage("")
	quotes.Add("quote 1")
	quotes.Add("quote 2")
	quotes.Add("quote 3")
	quotes.Add("quote 4")
	quotes.Add("quote 5")
	quotes.Add("quote 6")
	quotes.Add("quote 7")
	quotes.Add("quote 8")
	quotes.Add("quote 9")
	quotes.Add("quote 10")

	ctx = context.WithValue(ctx, "config", &cfg)
	ctx = context.WithValue(ctx, "clock", &mockClock)
	ctx = context.WithValue(ctx, "cache", requestCache)
	ctx = context.WithValue(ctx, "storage", quotes)
	ctx = context.WithValue(ctx, "logger", logger)

	t.Run("Handle close connection", func(t *testing.T) {
		input := fmt.Sprintf("%d%s", transport.CloseConnection, "|")
		msg, err := ProcessRequest(ctx, input, clientInfo)
		assert.Nil(t, msg)
		assert.Equal(t, nil, err)
	})

	t.Run("Invalid request", func(t *testing.T) {
		input := "||"
		msg, err := ProcessRequest(ctx, input, clientInfo)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "message malformed", err.Error())
	})

	t.Run("Unsupported operation", func(t *testing.T) {
		input := "100|"
		msg, err := ProcessRequest(ctx, input, clientInfo)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "unsupported operation", err.Error())
	})

	t.Run("can't recognize the type of message", func(t *testing.T) {
		input := "type|"
		msg, err := ProcessRequest(ctx, input, clientInfo)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "can't recognize the type of message", err.Error())
	})

	t.Run("Request challenge", func(t *testing.T) {
		input := fmt.Sprintf("%d|", transport.GetChallenge)
		msg, err := ProcessRequest(ctx, input, clientInfo)
		require.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, transport.Challenge, msg.Type)

		var solution pow.HashCashData
		err = json.Unmarshal([]byte(msg.Data), &solution)

		require.NoError(t, err)
		assert.Equal(t, 2, solution.ZerosCount)
		assert.Equal(t, clientInfo, solution.Resource)
		assert.NotEmpty(t, solution.Rand)
	})

	t.Run("Request resource without solution", func(t *testing.T) {
		input := fmt.Sprintf("%d|", transport.GetResource)
		msg, err := ProcessRequest(ctx, input, clientInfo)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.True(t, strings.Contains(err.Error(), "err unmarshal hashcash"))
	})

	t.Run("Request resource with wrong resource", func(t *testing.T) {
		hashcash := pow.HashCashData{
			Version:    1,
			ZerosCount: 4,
			Date:       time.Now().Unix(),
			Resource:   "client2",
			Rand:       base64.StdEncoding.EncodeToString([]byte(randKey)),
			Counter:    100,
		}
		marshaled, err := json.Marshal(hashcash)
		require.NoError(t, err)
		input := fmt.Sprintf("%d|%s", transport.GetResource, string(marshaled))
		msg, err := ProcessRequest(ctx, input, clientInfo)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "invalid hashcash resource", err.Error())
	})

	t.Run("Request resource with invalid solution and 0 counter", func(t *testing.T) {
		requestCache.Add(randKey, 100)

		hashcash := pow.HashCashData{
			Version:    1,
			ZerosCount: 10,
			Date:       time.Now().Unix(),
			Resource:   "client1",
			Rand:       base64.StdEncoding.EncodeToString([]byte(randKey)),
			Counter:    0,
		}
		marshaled, err := json.Marshal(hashcash)
		require.NoError(t, err)
		input := fmt.Sprintf("%d|%s", transport.GetResource, string(marshaled))
		msg, err := ProcessRequest(ctx, input, clientInfo)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "invalid hashcash resource", err.Error())
	})

	t.Run("Request resource with expired solution", func(t *testing.T) {
		mockClock.NowFunc = func() time.Time {
			return time.Date(1985, 7, 16, 0, 0, 0, 0, time.UTC)
		}
		requestCache.Add(randKey, 100)

		mockClock.NowFunc = func() time.Time {
			return time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC)
		}
		hashcash := pow.HashCashData{
			Version:    1,
			ZerosCount: 10,
			Date:       time.Date(2024, 1, 14, 1, 0, 0, 0, time.UTC).Unix(),
			Resource:   clientInfo,
			Rand:       base64.StdEncoding.EncodeToString([]byte(randKey)),
			Counter:    100,
		}
		marshaled, err := json.Marshal(hashcash)
		require.NoError(t, err)
		input := fmt.Sprintf("%d|%s", transport.GetResource, string(marshaled))
		msg, err := ProcessRequest(ctx, input, clientInfo)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "challenge expired or not sent", err.Error())
	})

	t.Run("Request resource with correct solution", func(t *testing.T) {
		mockClock.NowFunc = func() time.Time {
			return time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC)
		}
		err := requestCache.Add(randKey, 200)
		assert.NoError(t, err)

		mockClock.NowFunc = func() time.Time {
			return time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC)
		}

		date := time.Date(2024, 1, 14, 0, 0, 10, 0, time.UTC)
		hashcash := pow.HashCashData{
			Version:    1,
			ZerosCount: 1,
			Date:       date.Unix(),
			Resource:   clientInfo,
			Rand:       base64.StdEncoding.EncodeToString([]byte(randKey)),
			Counter:    14,
		}
		marshaled, err := json.Marshal(hashcash)
		require.NoError(t, err)
		input := fmt.Sprintf("%d|%s", transport.GetResource, string(marshaled))
		msg, err := ProcessRequest(ctx, input, clientInfo)
		require.NoError(t, err)
		assert.NotNil(t, msg)

		exists, err := requestCache.Contains(randKey)
		assert.NoError(t, err)
		assert.False(t, exists)

		msg, err = ProcessRequest(ctx, input, clientInfo)
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "challenge expired or not sent", err.Error())
	})
}
