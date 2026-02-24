package torrent

import (
	"strconv"
	"fmt"
	"os"
	"io"
	"strings"
	"log/slog"
	"kbit/internal/net"
	"kbit/internal/logger"
	"crypto/sha1"
	"kbit/pkg/types"
)

func ParseTorrentFile(file *os.File) (types.TorrentFile, error) {
	var torrent types.TorrentFile
	logger.Log.Info("parsing torrent file")

	data, err := io.ReadAll(file)
	if err != nil {
		return torrent, err
	}

	logger.Log.Debug("file read complete", slog.Int("bytes", len(data)))

	value, err := Decode(string(data))
	if err != nil {
		return torrent, err
	}

	root, ok := value.(types.BencodeDict)
	if !ok {
		err := fmt.Errorf("expected BencodeDict at root, got %T", value)
		return torrent, err
	}

	info, ok := root["info"].(types.BencodeDict)
	if !ok {
		err := fmt.Errorf("expected to find info at root")
		return torrent, err
	}

	name, ok := info["name"].(types.BencodeString)
	if !ok {
		err := fmt.Errorf("expected to find name at info")
		return torrent, err
	}

	torrent.Name = string(name)

	logger.Log.Info("torrent metadata",
		slog.String("name", torrent.Name),
	)

	privateStr, ok := info["private"].(types.BencodeString)
	if ok {
		private, err := strconv.ParseBool(string(privateStr))
		if err == nil {
			torrent.Private = private
		} else {
			logger.Log.Warn("invalid private flag",
				slog.String("value", string(privateStr)),
			)
		}
	}

	length, ok := info["length"].(types.BencodeInt)
	if ok {
		torrent.Length = int64(length)
		logger.Log.Debug("single file torrent",
			slog.Int64("length", torrent.Length),
		)
	} else {
		files, ok := info["files"].(types.BencodeList)
		if !ok {
			err := fmt.Errorf("torrent missing length and files")
			return torrent, err
		}

		var total int64
		for _, file := range files {
			length, ok := (file.(types.BencodeDict))["length"].(types.BencodeInt)
			if ok {
				total += int64(length)
			}
		}
		torrent.Length = total

		logger.Log.Debug("multi file torrent",
			slog.Int64("total_length", total),
		)
	}

	infoEncoded, err := Encode(info)
	if err != nil {
		return torrent, err
	}

	hasher := sha1.New()
	hasher.Write([]byte(infoEncoded))
	torrent.InfoHash = hasher.Sum(nil)

	logger.Log.Info("infohash generated",
		slog.String("infohash", fmt.Sprintf("%x", torrent.InfoHash)),
	)

	announce, _ := root["announce"].(types.BencodeString)
	if strings.HasPrefix(string(announce), "https://") || strings.HasPrefix(string(announce), "http://") {
		torrent.TrackerURL = string(announce)
		temp := make(types.HashSet[string], 1)
		temp[torrent.TrackerURL] = struct{}{}

		net.DiscoverPeers(&torrent, &temp)
	} else {
		logger.Log.Warn("torrent file main tracker url does not http/https protocol")
	}



	announceList, ok := root["announce-list"].(types.BencodeList)
	if ok {
		torrent.Trackers = make(types.HashSet[string])

		// WHY TORRENT JUST WHY
		for _, v1 := range announceList {
			for _, v2 := range (v1.(types.BencodeList)) {
				url := string(v2.(types.BencodeString))
				if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
					continue
				}

				torrent.Trackers[url] = struct{}{}
			}
		}

		net.DiscoverPeers(&torrent, &torrent.Trackers)
	}

	return torrent, nil
}
