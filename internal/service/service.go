package service

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/cache"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/clock"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/config"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/pow"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/storage"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/transport"
	"io"
	"log/slog"
	"math/rand"
	"net"
)

func Run(ctx context.Context, address string) error {
	logger := ctx.Value("logger").(*slog.Logger)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			logger.Error("close listener failed", "error", err)
		}
	}(listener)

	logger.Info("listening", "address", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept connection fail: %w", err)
		}
		go handleConnection(ctx, conn)
	}
}

func handleConnection(ctx context.Context, conn net.Conn) {
	logger := ctx.Value("logger").(*slog.Logger)
	clientInfo := conn.RemoteAddr()

	logger.Info("new client", "client info", clientInfo)

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("close connection failed", "error", err)
		}
	}(conn)

	reader := bufio.NewReader(conn)

	for {
		req, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			logger.Warn("read data error", "error", err)
			return
		}

		msg, err := ProcessRequest(ctx, req, clientInfo.String())
		if err != nil {
			logger.Error("can't process request", "error", err)
			return
		}

		if msg != nil {
			err := sendMessage(*msg, conn)
			if err != nil {
				logger.Error("can't send message", "error", err)
			}
		}
	}
}

func ProcessRequest(ctx context.Context, msgStr string, clientInfo string) (*transport.Message, error) {
	request, err := transport.ParseMessage(msgStr)
	if err != nil {
		return nil, err
	}
	logger := ctx.Value("logger").(*slog.Logger)
	logger.Info("got request from client", "address", clientInfo, "request type", request.Type, "request data", request.Data)
	switch request.Type {
	case transport.CloseConnection:
		logger.Info("handle close connection")
		return handleCloseConnection(ctx)
	case transport.GetChallenge:
		logger.Info("handle GetChallenge")
		return handleGetChallenge(ctx, clientInfo)
	case transport.GetResource:
		logger.Info("handle GetResource")
		return handleGetResource(ctx, clientInfo, request)
	default:
		return nil, fmt.Errorf("unsupported operation")
	}
}

func handleCloseConnection(ctx context.Context) (*transport.Message, error) {
	logger := ctx.Value("logger").(*slog.Logger)
	logger.Info("client closes connection")
	return nil, nil
}

func handleGetChallenge(ctx context.Context, clientInfo string) (*transport.Message, error) {
	logger := ctx.Value("logger").(*slog.Logger)
	cfg := ctx.Value("config").(*config.Config)
	_clock := ctx.Value("clock").(clock.Clock)
	requestCache := ctx.Value("cache").(cache.Cache)

	logger.Debug("client %s requests challenge", "client info", clientInfo)

	date := _clock.Now()

	randUUID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("get uuid failed: %w", err)
	}
	randValue := randUUID.String()
	err = requestCache.Add(randValue, cfg.Pow.Expiration)
	if err != nil {
		return nil, fmt.Errorf("err add rand to cache: %w", err)
	}

	powChallenge := pow.HashCashData{
		Version:    1,
		ZerosCount: cfg.Pow.Complexity,
		Date:       date.Unix(),
		Resource:   clientInfo,
		Rand:       base64.StdEncoding.EncodeToString([]byte(randValue)),
		Counter:    0,
	}
	challengeData, err := json.Marshal(powChallenge)
	if err != nil {
		return nil, fmt.Errorf("err marshal hashcash: %v", err)
	}

	msg := transport.Message{
		Type: transport.Challenge,
		Data: string(challengeData),
	}

	return &msg, nil
}

func handleGetResource(ctx context.Context, clientInfo string, request *transport.Message) (*transport.Message, error) {
	logger := ctx.Value("logger").(*slog.Logger)
	logger.Debug("client requests resource with payload", "client", clientInfo, "data", request.Data)
	var powResult pow.HashCashData

	err := json.Unmarshal([]byte(request.Data), &powResult)
	if err != nil {
		return nil, fmt.Errorf("err unmarshal hashcash: %w", err)
	}

	if powResult.Version == 1 && powResult.Resource != clientInfo {
		return nil, fmt.Errorf("invalid hashcash resource")
	}

	err = verifySolution(ctx, powResult)
	if err != nil {
		return nil, err
	}

	logger.Debug("client successfully solves challenge", "client", clientInfo, "data", request.Data)

	msg := transport.Message{
		Type: transport.Resource,
		Data: getRandomQuote(ctx),
	}

	return &msg, err
}

func verifySolution(ctx context.Context, powResult pow.HashCashData) error {
	cfg := ctx.Value("config").(*config.Config)
	_clock := ctx.Value("clock").(clock.Clock)
	requestCache := ctx.Value("cache").(cache.Cache)

	randValueBytes, err := base64.StdEncoding.DecodeString(powResult.Rand)
	if err != nil {
		return fmt.Errorf("err decode rand: %w", err)
	}

	randValue := string(randValueBytes)
	if err != nil {
		return fmt.Errorf("err decode rand: %w", err)
	}

	exists, err := requestCache.Contains(randValue)
	if err != nil {
		return fmt.Errorf("err get rand from cache: %w", err)
	}
	if !exists {
		return fmt.Errorf("challenge expired or not sent")
	}

	if _clock.Now().Unix()-powResult.Date > cfg.Pow.Expiration {
		return fmt.Errorf("challenge expired")
	}

	maxTries := powResult.Counter
	if maxTries == 0 {
		maxTries = 1
	}

	_, err = powResult.Compute(maxTries)
	if err != nil {
		return fmt.Errorf("wrong solution")
	}
	err = requestCache.Delete(randValue)
	if err != nil {
		return fmt.Errorf("invalid hashcash")
	}
	return nil
}

func getRandomQuote(ctx context.Context) string {
	quoteStorage := ctx.Value("storage").(storage.Storage)
	numberOfQuotes := quoteStorage.Size()
	return quoteStorage.Get(rand.Intn(numberOfQuotes))
}

func sendMessage(msg transport.Message, conn net.Conn) error {
	msgStr := fmt.Sprintf("%s\n", msg.ToString())
	_, err := conn.Write([]byte(msgStr))
	return err
}
