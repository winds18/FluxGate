package observability

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func NewLogger(logDir string) (*slog.Logger, func(), error) {
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return nil, func() {}, err
	}

	logPath := filepath.Join(logDir, "fluxgate.log")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, func() {}, err
	}

	writer := io.MultiWriter(os.Stdout, file)
	logger := slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return logger, func() { _ = file.Close() }, nil
}

func NewRequestID() string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(bytes[:])
}
