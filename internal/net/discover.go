package net

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"log/slog"
	"crypto/rand"
	"encoding/hex"

	"kbit/internal/logger"
	"kbit/pkg/types"
)

func DiscoverPeers(torrent *types.TorrentFile, urls *types.HashSet[string]) error {
	if ((*torrent).Private || len(*urls) == 1) {
		peers := bruteForce(torrent, urls)
		torrent.Peers = peers
		return nil
	}

	peers := fetchConcurent(torrent, urls)
	torrent.Peers = peers
	return nil
}

func fetchConcurent(torrent *types.TorrentFile, urls *types.HashSet[string]) types.HashSet[string] {
	peers := make(types.HashSet[string])

	var wg sync.WaitGroup
	var mu sync.Mutex

	limit := make(chan struct{}, 10)

	for trackerURL := range *urls {
		wg.Add(1)

		go func(tracker string) {
			defer wg.Done()

			limit <- struct{}{}
			defer func() { <-limit }()

			reqURL, err := buildTrackerURL(torrent, tracker)
			if err != nil {
				logger.Log.Error("invalid tracker URL",
					slog.String("tracker", tracker),
					slog.String("error", err.Error()),
				)
				return
			}

			logger.Log.Info("querying tracker",
				slog.String("tracker", tracker),
				slog.String("infohash", fmt.Sprintf("%x", torrent.InfoHash)),
			)

			resp, err := http.Get(reqURL)
			if err != nil {
				logger.Log.Warn("failed to reach tracker",
					slog.String("tracker", tracker),
					slog.String("error", err.Error()),
				)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Log.Warn("failed to read tracker response",
					slog.String("tracker", tracker),
					slog.String("error", err.Error()),
				)
				return
			}

			temp := make(types.HashSet[string])
			extractPeers(body, temp)

			mu.Lock()
			for p := range temp {
				peers[p] = struct{}{}
			}

			mu.Unlock()
		}(trackerURL)
	}

	wg.Wait()
	return peers
}

func bruteForce(torrent *types.TorrentFile, urls *types.HashSet[string]) types.HashSet[string] {
	peers := make(types.HashSet[string])
	for trackerURL := range *urls {
		reqURL, err := buildTrackerURL(torrent, trackerURL)
		if err != nil {
			logger.Log.Error("invalid tracker URL",
				slog.String("tracker", trackerURL),
				slog.String("error", err.Error()),
			)
			continue
		}

		logger.Log.Info("querying tracker",
			slog.String("tracker", trackerURL),
			slog.String("infohash", fmt.Sprintf("%x", torrent.InfoHash)),
		)

		resp, err := http.Get(reqURL)
		if err != nil {
			logger.Log.Warn("failed to reach tracker",
				slog.String("tracker", trackerURL),
				slog.String("error", err.Error()),
			)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logger.Log.Warn("failed to read tracker response",
				slog.String("tracker", trackerURL),
				slog.String("error", err.Error()),
			)
			continue
		}

		extractPeers(body, peers)
	}

	return peers
}

func generatePeerID() string {
	random := make([]byte, 12)
	rand.Read(random)
	return "-GT0001-" + hex.EncodeToString(random)[:12]
}

func buildTrackerURL(torrent *types.TorrentFile, tracker string) (string, error) {
	base, err := url.Parse(tracker)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Set("info_hash", string(torrent.InfoHash))
	params.Set("peer_id", generatePeerID())
	params.Set("port", "6881")
	params.Set("uploaded", "0")
	params.Set("downloaded", "0")
	params.Set("left", strconv.Itoa(int(torrent.Length)))
	params.Set("compact", "1")

	base.RawQuery = params.Encode()
	return base.String(), nil
}

func extractPeers(body []byte, peers types.HashSet[string]) {
	// Look for: "5:peers"
	key := []byte("5:peers")
	idx := bytes.Index(body, key)
	if idx == -1 {
		return
	}

	// Move to start of peer string length
	start := idx + len(key)

	// Read length prefix (e.g., 12:)
	colon := bytes.IndexByte(body[start:], ':')
	if colon == -1 {
		return
	}

	lengthStr := string(body[start : start+colon])
	var peerLen int
	fmt.Sscanf(lengthStr, "%d", &peerLen)

	peerStart := start + colon + 1
	peerData := body[peerStart : peerStart+peerLen]

	// Compact peer list: 6 bytes per peer
	for i := 0; i+6 <= len(peerData); i += 6 {
		ip := net.IP(peerData[i : i+4])
		port := binary.BigEndian.Uint16(peerData[i+4 : i+6])

		addr := fmt.Sprintf("%s:%d", ip.String(), port)
		peers[addr] = struct{}{}
	}
}
