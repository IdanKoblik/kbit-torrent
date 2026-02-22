package torrent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTorrentFile_SingleFile(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "test.torrent")

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer file.Close()

	torrent, err := ParseTorrentFile(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if torrent.Name == "" {
		t.Error("expected torrent name to be set")
	}

	if torrent.Length == 0 {
		t.Error("expected torrent length to be greater than 0")
	}

	if len(torrent.InfoHash) != 20 {
		t.Errorf("expected infohash length 20, got %d", len(torrent.InfoHash))
	}
}

func TestParseTorrentFile_InvalidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid*.torrent")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("not a torrent")
	tmpFile.Close()

	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	_, err = ParseTorrentFile(file)
	if err == nil {
		t.Error("expected error for invalid torrent file")
	}
}

func TestParseTorrentFile_MissingInfo(t *testing.T) {
	content := "d3:foo3:bare"

	tmpFile, err := os.CreateTemp("", "missinginfo*.torrent")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(content)
	tmpFile.Close()

	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	_, err = ParseTorrentFile(file)
	if err == nil {
		t.Error("expected error for missing info dictionary")
	}
}
