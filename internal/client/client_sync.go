package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/config"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/pow"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/transport"
	"io"
	"log/slog"
	"net"
	"time"
)

func RunSync(ctx context.Context, address string) error {
	logger := ctx.Value("logger").(*slog.Logger)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Error("dial server error", "error", err)
		return err
	}
	logger.Info("connected", "address", address)

	defer closeConnection(conn, logger)

	for {
		message, err := HandleConnection(ctx, conn, conn)
		if err != nil {
			logger.Error("handle connection failed", "error", err)
			return err
		}
		logger.Info("got quote", "quote", message)
		time.Sleep(5 * time.Second)
	}
}

func HandleConnection(ctx context.Context, connReader io.Reader, connWriter io.Writer) (string, error) {
	logger := ctx.Value("logger").(*slog.Logger)
	reader := bufio.NewReader(connReader)

	err := sendMsg(transport.Message{
		Type: transport.GetChallenge,
	}, connWriter)
	if err != nil {
		return "", fmt.Errorf("err send request: %w", err)
	}

	msgStr, err := receiveMsg(reader)
	if err != nil {
		return "", fmt.Errorf("err read msg: %w", err)
	}
	msg, err := transport.ParseMessage(msgStr)
	if err != nil {
		return "", fmt.Errorf("err parse msg: %w", err)
	}

	var challengeData pow.HashCashData
	err = json.Unmarshal([]byte(msg.Data), &challengeData)
	if err != nil {
		return "", fmt.Errorf("err parse hashcash: %w", err)
	}
	logger.Debug("got challenge", "data", challengeData)

	cfg := ctx.Value("config").(*config.Config)
	solution, err := challengeData.Compute(cfg.Pow.MaxIterations)
	if err != nil {
		return "", fmt.Errorf("err compute hashcash: %w", err)
	}
	logger.Debug("solution computed:", "solution", solution)

	err = sendMsg(transport.Message{
		Type: transport.GetResource,
		Data: solution,
	}, connWriter)
	if err != nil {
		return "", fmt.Errorf("err send request: %w", err)
	}

	logger.Debug("challenge sent to server")

	msgStr, err = receiveMsg(reader)
	if err != nil {
		return "", fmt.Errorf("err read msg: %w", err)
	}
	msg, err = transport.ParseMessage(msgStr)
	if err != nil {
		return "", fmt.Errorf("err parse msg: %w", err)
	}
	return msg.Data, nil
}
