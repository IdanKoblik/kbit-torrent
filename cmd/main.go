package main

import (
	"fmt"
	"os"
	"kbit/internal/torrent"
)

func main() {
	file, err := os.Open("test.torrent")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	torrent.ParseTorrentFile(file)
}
