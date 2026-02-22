package torrent

import (
	"fmt"
	"os"
	"io"
	"kbit/pkg/types"
)

func ParseTorrentFile(file *os.File) (types.TorrentFile, error) {
	var torrent types.TorrentFile

	data, err := io.ReadAll(file)
	if err != nil {
		return torrent, err
	}

	value, err := Decode(string(data[:]))
	if err != nil {
		return torrent, err
	}

	root, ok := value.(types.BencodeDict)
	if !ok {
		return torrent, fmt.Errorf("expected BencodeDict at root, got %T", value)
	}

	root.Print(0)

	return torrent, nil
}
