package cmd

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

func TestFindCommand_Parse(t *testing.T) {
	cmd, err := FindCommand("parse")
	if err != nil {
		t.Fatalf("expected no error for 'parse', got: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected a non-nil command")
	}
}

func TestFindCommand_UnknownCommand(t *testing.T) {
	_, err := FindCommand("notacommand")
	if err == nil {
		t.Error("expected error for unknown command name")
	}
}

func TestFindCommand_EmptyName(t *testing.T) {
	_, err := FindCommand("")
	if err == nil {
		t.Error("expected error for empty command name")
	}
}

func TestParseCommand_FileNotFound(t *testing.T) {
	cmd := &ParseCommand{}
	err := cmd.Run("/nonexistent/path/no.torrent")
	if err == nil {
		t.Error("expected error when file does not exist")
	}
}

func TestParseCommand_InvalidTorrent(t *testing.T) {
	f, err := os.CreateTemp("", "invalid*.torrent")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("this is not bencode")
	f.Close()

	cmd := &ParseCommand{}
	if err := cmd.Run(f.Name()); err == nil {
		t.Error("expected error for non-bencode content")
	}
}

func TestParseCommand_ValidTorrent_UDPAnnounce(t *testing.T) {
	content := "d8:announce28:udp://tracker.example.com:804:infod6:lengthi1024e4:name8:testfileee"

	f, err := os.CreateTemp("", "valid*.torrent")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	cmd := &ParseCommand{}
	if err := cmd.Run(f.Name()); err != nil {
		t.Errorf("expected no error for valid torrent, got: %v", err)
	}
}
