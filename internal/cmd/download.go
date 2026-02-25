package cmd

import (
	"fmt"
	"os"

	"kbit/internal/net"
	"kbit/internal/torrent"
)

type DownloadCommand struct{}

func (c *DownloadCommand) Run(arg string) error {
	file, err := os.Open(arg)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", arg, err)
	}
	defer file.Close()

	t, err := torrent.ParseTorrentFile(file)
	if err != nil {
		return err
	}

	return net.Download(&t)
}
