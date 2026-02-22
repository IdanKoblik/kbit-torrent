package cmd

import (
	"os"
	"kbit/internal/torrent"
	"fmt"
)

type ParseCommand struct {
	File string
}

func (c *ParseCommand) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide a file")
	}

	c.File = args[0]
	file, err := os.Open(c.File)
	if err != nil {
		return fmt.Errorf("File %s does not exists", c.File)
	}

	defer file.Close()

	torrent, err := torrent.ParseTorrentFile(file)
	if err != nil {
		return err
	}

	fmt.Print("\033[H\033[2J") // move cursor to top-left + clear screen

	fmt.Println("")
	fmt.Println("===== SUMMARY =====")
	fmt.Printf("Name: %s\n", torrent.Name)
	fmt.Printf("Private: %t\n", torrent.Private)
	fmt.Printf("Info hash: %x\n", string(torrent.InfoHash))
	fmt.Printf("Length: %d\n", torrent.Length)

	return nil
}
