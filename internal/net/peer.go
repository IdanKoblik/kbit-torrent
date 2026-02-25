package net

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	MsgChoke         uint8 = 0
	MsgUnchoke       uint8 = 1
	MsgInterested    uint8 = 2
	MsgNotInterested uint8 = 3
	MsgHave          uint8 = 4
	MsgBitfield      uint8 = 5
	MsgRequest       uint8 = 6
	MsgPiece         uint8 = 7
	MsgCancel        uint8 = 8

	BlockSize = 16 * 1024
)

type PeerMsg struct {
	ID      uint8
	Payload []byte
}

type PeerConn struct {
	conn     net.Conn
	Addr     string
	Bitfield []byte
}

func NewPeerConn(conn net.Conn, addr string) *PeerConn {
	return &PeerConn{conn: conn, Addr: addr}
}

func (p *PeerConn) SendMsg(id uint8, payload []byte) error {
	length := uint32(1 + len(payload))
	buf := make([]byte, 4+1+len(payload))
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = id
	copy(buf[5:], payload)
	_, err := p.conn.Write(buf)
	return err
}

func (p *PeerConn) ReadMsg() (*PeerMsg, error) {
	var lengthBuf [4]byte
	if _, err := io.ReadFull(p.conn, lengthBuf[:]); err != nil {
		return nil, fmt.Errorf("reading message length: %w", err)
	}
	length := binary.BigEndian.Uint32(lengthBuf[:])
	if length == 0 {
		// keep-alive
		return nil, nil
	}

	msgBuf := make([]byte, length)
	if _, err := io.ReadFull(p.conn, msgBuf); err != nil {
		return nil, fmt.Errorf("reading message body: %w", err)
	}
	return &PeerMsg{ID: msgBuf[0], Payload: msgBuf[1:]}, nil
}

func (p *PeerConn) SetDeadline(t time.Time) error {
	return p.conn.SetDeadline(t)
}

func (p *PeerConn) Close() error {
	return p.conn.Close()
}

func (p *PeerConn) HasPiece(i int) bool {
	byteIdx := i / 8
	bitIdx := 7 - (i % 8)
	if byteIdx >= len(p.Bitfield) {
		return false
	}
	return p.Bitfield[byteIdx]>>uint(bitIdx)&1 == 1
}

func buildRequestPayload(index, begin, length int) []byte {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return payload
}
