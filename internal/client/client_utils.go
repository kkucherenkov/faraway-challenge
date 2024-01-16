package client

import (
	"bufio"
	"fmt"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/transport"
	"io"
	"log/slog"
	"net"
)

// receiveMsg - read string message from connection
func receiveMsg(reader *bufio.Reader) (string, error) {
	return reader.ReadString('\n')
}

// sendMsg - send protocol message to connection
func sendMsg(msg transport.Message, conn io.Writer) error {
	msgStr := fmt.Sprintf("%s\n", msg.ToString())
	_, err := conn.Write([]byte(msgStr))
	return err
}

func closeConnection(conn net.Conn, logger *slog.Logger) {
	err := conn.Close()
	if err != nil {
		logger.Error("close connection failed", "error", err)
	}
}
