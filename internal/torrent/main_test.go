package torrent

import (
	"log/slog"
	"os"
	"testing"

	"kbit/internal/logger"
)

func TestMain(m *testing.M) {
	logger.Init(slog.LevelError, false)
	os.Exit(m.Run())
}
