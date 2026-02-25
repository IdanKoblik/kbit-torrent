package net

import (
	"fmt"
	"io"
	"net"
	"time"
)

func Handshake(addr string, infoHash []byte) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, err
	}

	if len(PeerID) != 20 {
		conn.Close()
		return nil, fmt.Errorf("peerID must be 20 bytes")
	}

	pstr := "BitTorrent protocol"
	handshake := make([]byte, 49+len(pstr))

	handshake[0] = byte(len(pstr))
	copy(handshake[1:], pstr)
	// 8 reserved bytes are already zero
	copy(handshake[1+len(pstr)+8:], infoHash)
	copy(handshake[1+len(pstr)+8+20:], PeerID)

	conn.SetDeadline(time.Now().Add(10 * time.Second))

	_, err = conn.Write(handshake)
	if err != nil {
		conn.Close()
		return nil, err
	}

	resp := make([]byte, 68)
	_, err = io.ReadFull(conn, resp)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if string(resp[1:20]) != pstr {
		conn.Close()
		return nil, fmt.Errorf("invalid protocol string")
	}

	if string(resp[28:48]) != string(infoHash) {
		conn.Close()
		return nil, fmt.Errorf("infohash mismatch")
	}

	return conn, nil
}
