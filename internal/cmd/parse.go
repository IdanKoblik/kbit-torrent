package cmd

import (
	"os"
	"kbit/internal/torrent"
	"fmt"
)

type ParseCommand struct {
	File string
}

func (c *ParseCommand) Run(arg string) error {
	c.File = arg
	file, err := os.Open(c.File)
	if err != nil {
		return fmt.Errorf("File %s does not exists", c.File)
	}

	defer file.Close()

	torrent, err := torrent.ParseTorrentFile(file)
	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("===== SUMMARY =====")
	fmt.Printf("Name: %s\n", torrent.Name)
	fmt.Printf("Private: %t\n", torrent.Private)
	fmt.Printf("Info hash: %x\n", string(torrent.InfoHash))
	fmt.Printf("Length: %d\n", torrent.Length)
	fmt.Println("")
	fmt.Printf("Tracker: %s\n", torrent.TrackerURL)
	for peer := range torrent.Peers {
		fmt.Println(peer)
	}

	return nil
}
