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

func RunAsync(ctx context.Context) error {
	logger := ctx.Value("logger").(*slog.Logger)
	cfg := ctx.Value("config").(*config.Config)
	serverAddress := fmt.Sprintf("%s:%s", cfg.Service.Host, cfg.Service.Port)
	for {
		// dial service to get challenge
		conn, err := net.Dial("tcp", serverAddress)
		if err != nil {
			logger.Error("dial service error", "error", err)
			return err
		}
		logger.Info("connected to get challenge", "address", serverAddress)

		challenge, err := GetChallenge(conn, conn, logger)
		if err != nil {
			logger.Error("Get challenge failed", "error", err)
			return err
		}
		closeConnection(conn, logger)

		solution, err := SolveChallenge(cfg, logger, challenge)
		if err != nil {
			logger.Error("Solve challenge failed", "error", err)
			return err
		}

		conn, err = net.Dial("tcp", serverAddress)
		if err != nil {
			logger.Error("dial service error", "error", err)
			return err
		}
		logger.Info("connected to get quote", "address", serverAddress)

		quote, err := GetQuote(conn, conn, logger, solution)
		if err != nil {
			logger.Error("Get quote failed", "error", err)
			return err
		}
		closeConnection(conn, logger)

		logger.Info("got quote", "quote", quote)
		time.Sleep(1 * time.Second)
	}
}

func GetChallenge(connReader io.Reader, connWriter io.Writer, logger *slog.Logger) (*pow.HashCashData, error) {
	reader := bufio.NewReader(connReader)

	err := sendMsg(transport.Message{
		Type: transport.GetChallenge,
	}, connWriter)
	if err != nil {
		return nil, fmt.Errorf("err send request: %w", err)
	}

	msgStr, err := receiveMsg(reader)
	if err != nil {
		return nil, fmt.Errorf("err read msg: %w", err)
	}
	msg, err := transport.ParseMessage(msgStr)
	if err != nil {
		return nil, fmt.Errorf("err parse msg: %w", err)
	}

	var challengeData pow.HashCashData
	err = json.Unmarshal([]byte(msg.Data), &challengeData)
	if err != nil {
		return nil, fmt.Errorf("err parse hashcash: %w", err)
	}
	logger.Debug("got challenge")

	err = sendMsg(transport.Message{
		Type: transport.CloseConnection,
	}, connWriter)
	if err != nil {
		logger.Warn("can't close connection", "error", err)
	}
	return &challengeData, nil
}

func SolveChallenge(cfg *config.Config, logger *slog.Logger, challenge *pow.HashCashData) (string, error) {
	// to prevent client info check on service side
	challenge.Version = 2
	solution, err := challenge.Compute(cfg.Pow.MaxIterations)
	if err != nil {
		return "", fmt.Errorf("err compute hashcash: %w", err)
	}
	logger.Debug("solution computed:", "solution", solution)
	return solution, err
}

func GetQuote(connReader io.Reader, connWriter io.Writer, logger *slog.Logger, solution string) (string, error) {
	reader := bufio.NewReader(connReader)
	err := sendMsg(transport.Message{
		Type: transport.GetResource,
		Data: solution,
	}, connWriter)
	if err != nil {
		return "", fmt.Errorf("err send request: %w", err)
	}

	logger.Debug("challenge sent to server")

	msgStr, err := receiveMsg(reader)
	if err != nil {
		return "", fmt.Errorf("err read msg: %w", err)
	}
	msg, err := transport.ParseMessage(msgStr)
	if err != nil {
		return "", fmt.Errorf("err parse msg: %w", err)
	}

	err = sendMsg(transport.Message{
		Type: transport.CloseConnection,
	}, connWriter)
	if err != nil {
		logger.Warn("can't close connection", "error", err)
	}
	return msg.Data, nil
}
