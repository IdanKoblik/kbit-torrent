package net

import (
	"io"
	"net"
	"testing"
)

// buildHandshakeResponse constructs the 68-byte handshake that a peer sends back.
func buildHandshakeResponse(infoHash []byte) []byte {
	pstr := "BitTorrent protocol"
	resp := make([]byte, 68)
	resp[0] = byte(len(pstr))
	copy(resp[1:20], pstr)
	// bytes 20-27: reserved (zero)
	copy(resp[28:48], infoHash)
	// bytes 48-67: peer ID (arbitrary for tests)
	copy(resp[48:68], "-GT0001-TESTPEERID-")
	return resp
}

// startMockPeer starts a local TCP listener that performs a valid BitTorrent
// handshake and returns the listener address.
func startMockPeer(t *testing.T, infoHash []byte) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not listen: %v", err)
	}

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		defer ln.Close()

		// Drain the client handshake (68 bytes).
		buf := make([]byte, 68)
		io.ReadFull(conn, buf) //nolint:errcheck

		// Respond with a valid handshake.
		conn.Write(buildHandshakeResponse(infoHash)) //nolint:errcheck
	}()

	return ln.Addr().String()
}

func TestHandshake_Success(t *testing.T) {
	infoHash := make([]byte, 20)
	copy(infoHash, "01234567890123456789")

	PeerID = []byte("-GT0001-LOCALPEERID-") // must be 20 bytes

	addr := startMockPeer(t, infoHash)

	conn, err := Handshake(addr, infoHash)
	if err != nil {
		t.Fatalf("expected successful handshake, got: %v", err)
	}
	conn.Close()
}

func TestHandshake_ConnectionRefused(t *testing.T) {
	PeerID = []byte("-GT0001-LOCALPEERID-")

	// Port 1 is almost always refused.
	_, err := Handshake("127.0.0.1:1", make([]byte, 20))
	if err == nil {
		t.Error("expected error for refused connection")
	}
}

func TestHandshake_InfoHashMismatch(t *testing.T) {
	infoHash := make([]byte, 20)
	copy(infoHash, "01234567890123456789")
	wrongHash := make([]byte, 20)
	copy(wrongHash, "AAAAAAAAAAAAAAAAAAAAA")

	PeerID = []byte("-GT0001-LOCALPEERID-")

	// Peer sends back the wrong infohash.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not listen: %v", err)
	}
	go func() {
		conn, _ := ln.Accept()
		defer conn.Close()
		defer ln.Close()
		buf := make([]byte, 68)
		io.ReadFull(conn, buf) //nolint:errcheck
		conn.Write(buildHandshakeResponse(wrongHash)) //nolint:errcheck
	}()

	_, err = Handshake(ln.Addr().String(), infoHash)
	if err == nil {
		t.Error("expected error for infohash mismatch")
	}
}

func TestHandshake_InvalidProtocolString(t *testing.T) {
	infoHash := make([]byte, 20)
	PeerID = []byte("-GT0001-LOCALPEERID-")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not listen: %v", err)
	}
	go func() {
		conn, _ := ln.Accept()
		defer conn.Close()
		defer ln.Close()
		buf := make([]byte, 68)
		io.ReadFull(conn, buf) //nolint:errcheck

		resp := make([]byte, 68)
		resp[0] = 19
		copy(resp[1:20], "NotBitTorrentProtoc") // wrong protocol string
		copy(resp[28:48], infoHash)
		conn.Write(resp) //nolint:errcheck
	}()

	_, err = Handshake(ln.Addr().String(), infoHash)
	if err == nil {
		t.Error("expected error for invalid protocol string")
	}
}

func TestHandshake_PeerIDTooShort(t *testing.T) {
	PeerID = []byte("short") // not 20 bytes

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not listen: %v", err)
	}
	defer ln.Close()

	_, err = Handshake(ln.Addr().String(), make([]byte, 20))
	if err == nil {
		t.Error("expected error for peer ID shorter than 20 bytes")
	}
}
