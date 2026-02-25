package cmd

import (
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"

	"kbit/internal/logger"
	knet "kbit/internal/net"
)

func TestMain(m *testing.M) {
	logger.Init(slog.LevelError, false)
	os.Exit(m.Run())
}

// FindCommand
func TestFindCommand_Parse(t *testing.T) {
	cmd, err := FindCommand("parse")
	if err != nil {
		t.Fatalf("expected no error for 'parse', got: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected a non-nil command")
	}
}

func TestFindCommand_Handshake(t *testing.T) {
	cmd, err := FindCommand("handshake")
	if err != nil {
		t.Fatalf("expected no error for 'handshake', got: %v", err)
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

// ParseCommand

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

// HandshakeCommand helpers

// writeTempTorrent creates a temp file with the given content and registers
// cleanup. Returns the file path.
func writeTempTorrent(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "*.torrent")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	f.WriteString(content)
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

// startEchoPeer starts a TCP listener that reads the client's handshake and
// echoes the same infohash back, making the handshake always succeed regardless
// of the actual infohash value.
func startEchoPeer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		defer ln.Close()

		// Drain the 68-byte client handshake.
		buf := make([]byte, 68)
		io.ReadFull(conn, buf) //nolint:errcheck

		// Build a valid response echoing the received infohash (bytes 28-47).
		pstr := "BitTorrent protocol"
		resp := make([]byte, 68)
		resp[0] = byte(len(pstr))
		copy(resp[1:20], pstr)
		copy(resp[28:48], buf[28:48]) // echo infohash back
		copy(resp[48:68], "-GT0001-TESTPEERID--")
		conn.Write(resp) //nolint:errcheck
	}()
	return ln.Addr().String()
}

const validTorrentContent = "d8:announce28:udp://tracker.example.com:804:infod6:lengthi1024e4:name8:testfileee"

// HandshakeCommand tests

func TestHandshakeCommand_FileNotFound(t *testing.T) {
	cmd := &HandshakeCommand{In: strings.NewReader("")}
	if err := cmd.Run("/nonexistent/path/no.torrent"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestHandshakeCommand_InvalidTorrent(t *testing.T) {
	path := writeTempTorrent(t, "this is not bencode")
	cmd := &HandshakeCommand{In: strings.NewReader("")}
	if err := cmd.Run(path); err == nil {
		t.Error("expected error for invalid torrent content")
	}
}

func TestHandshakeCommand_EmptyPeerAddress(t *testing.T) {
	path := writeTempTorrent(t, validTorrentContent)
	// Simulate user pressing Enter without typing an address.
	cmd := &HandshakeCommand{In: strings.NewReader("\n")}
	if err := cmd.Run(path); err == nil {
		t.Error("expected error for empty peer address")
	}
}

func TestHandshakeCommand_ConnectionRefused(t *testing.T) {
	knet.PeerID = []byte("-GT0001-LOCALPEERID-")
	path := writeTempTorrent(t, validTorrentContent)
	// Port 1 is essentially always refused.
	cmd := &HandshakeCommand{In: strings.NewReader("127.0.0.1:1\n")}
	if err := cmd.Run(path); err == nil {
		t.Error("expected error for refused connection")
	}
}

func TestHandshakeCommand_Success(t *testing.T) {
	knet.PeerID = []byte("-GT0001-LOCALPEERID-")
	path := writeTempTorrent(t, validTorrentContent)
	addr := startEchoPeer(t)

	cmd := &HandshakeCommand{In: strings.NewReader(addr + "\n")}
	if err := cmd.Run(path); err != nil {
		t.Errorf("expected successful handshake, got: %v", err)
	}
}
