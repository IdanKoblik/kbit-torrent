package torrent

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempTorrent(t *testing.T, content string) *os.File {
	t.Helper()
	tmp, err := os.CreateTemp("", "*.torrent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	f, err := os.Open(tmp.Name())
	if err != nil {
		os.Remove(tmp.Name())
		t.Fatal(err)
	}
	t.Cleanup(func() {
		f.Close()
		os.Remove(tmp.Name())
	})
	return f
}

func TestParseTorrentFile_SingleFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

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

func TestParseTorrentFile_UDPAnnounce(t *testing.T) {
	content := "d8:announce28:udp://tracker.example.com:804:infod6:lengthi1024e4:name8:testfileee"
	file := writeTempTorrent(t, content)

	torrent, err := ParseTorrentFile(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if torrent.Name != "testfile" {
		t.Errorf("expected name 'testfile', got %q", torrent.Name)
	}
	if torrent.Length != 1024 {
		t.Errorf("expected length 1024, got %d", torrent.Length)
	}
	if len(torrent.InfoHash) != 20 {
		t.Errorf("expected infohash length 20, got %d", len(torrent.InfoHash))
	}
	if torrent.Private {
		t.Error("expected Private to be false")
	}
}

func TestParseTorrentFile_MultiFile(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "multi-file.torrent")
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer file.Close()

	torrent, err := ParseTorrentFile(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if torrent.Name != "test-multi" {
		t.Errorf("expected name 'test-multi', got %q", torrent.Name)
	}
	// Total of the two files: 100 + 200 = 300
	if torrent.Length != 300 {
		t.Errorf("expected total length 300, got %d", torrent.Length)
	}
	if len(torrent.InfoHash) != 20 {
		t.Errorf("expected infohash length 20, got %d", len(torrent.InfoHash))
	}
}

func TestParseTorrentFile_PrivateTorrent(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "private.torrent")
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer file.Close()

	torrent, err := ParseTorrentFile(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !torrent.Private {
		t.Error("expected Private to be true")
	}
	if torrent.Name != "secret" {
		t.Errorf("expected name 'secret', got %q", torrent.Name)
	}
}

func TestParseTorrentFile_MissingLengthAndFiles(t *testing.T) {
	content := "d4:infod4:name4:testee"
	file := writeTempTorrent(t, content)

	_, err := ParseTorrentFile(file)
	if err == nil {
		t.Error("expected error when info has neither length nor files")
	}
}

func TestParseTorrentFile_NoAnnounce(t *testing.T) {
	content := "d4:infod6:lengthi512e4:name6:simpleee"
	file := writeTempTorrent(t, content)

	torrent, err := ParseTorrentFile(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if torrent.Name != "simple" {
		t.Errorf("expected name 'simple', got %q", torrent.Name)
	}
	if torrent.TrackerURL != "" {
		t.Errorf("expected empty tracker URL, got %q", torrent.TrackerURL)
	}
	if torrent.Length != 512 {
		t.Errorf("expected length 512, got %d", torrent.Length)
	}
}
