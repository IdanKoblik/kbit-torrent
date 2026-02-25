package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"kbit/internal/net"
	"kbit/internal/torrent"
)

type HandshakeCommand struct {
	In io.Reader // overridden in tests; defaults to os.Stdin
}

func (c *HandshakeCommand) Run(arg string) error {
	file, err := os.Open(arg)
	if err != nil {
		return fmt.Errorf("file %s does not exist", arg)
	}
	defer file.Close()

	t, err := torrent.ParseTorrentFile(file)
	if err != nil {
		return err
	}

	in := c.In
	if in == nil {
		in = os.Stdin
	}

	fmt.Print("Enter peer address (host:port): ")
	scanner := bufio.NewScanner(in)
	scanner.Scan()
	addr := strings.TrimSpace(scanner.Text())
	if addr == "" {
		return fmt.Errorf("peer address cannot be empty")
	}

	conn, err := net.Handshake(addr, t.InfoHash)
	if err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}
	conn.Close()

	fmt.Printf("Handshake successful with %s\n", addr)
	return nil
}
