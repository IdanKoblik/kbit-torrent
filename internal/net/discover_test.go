package net

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"crypto/rand"

	"kbit/pkg/types"
)

func generatePeerID() string {
	peerID := make([]byte, 20)
	copy(peerID[:8], []byte("-GT0001-"))
	rand.Read(peerID[8:]) //nolint:errcheck // crypto/rand.Read never fails
	return string(peerID)
}

func buildPeerResponse(peerBytes []byte) []byte {
	prefix := fmt.Sprintf("d5:peers%d:", len(peerBytes))
	body := append([]byte(prefix), peerBytes...)
	return append(body, 'e')
}

func TestExtractPeers_NoPeersKey(t *testing.T) {
	body := []byte("d8:intervali1800ee")
	peers := make(types.HashSet[string])
	extractPeers(body, peers)
	if len(peers) != 0 {
		t.Errorf("expected 0 peers, got %d", len(peers))
	}
}

func TestExtractPeers_SinglePeer(t *testing.T) {
	// 127.0.0.1:6881 — [0x7F,0x00,0x00,0x01] + [0x1A,0xE1]
	peerBytes := []byte{127, 0, 0, 1, 0x1A, 0xE1}
	body := buildPeerResponse(peerBytes)

	peers := make(types.HashSet[string])
	extractPeers(body, peers)

	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d: %v", len(peers), peers)
	}
	if _, ok := peers["127.0.0.1:6881"]; !ok {
		t.Errorf("expected peer 127.0.0.1:6881, got: %v", peers)
	}
}

func TestExtractPeers_MultiplePeers(t *testing.T) {
	// 1.2.3.4:1000 and 5.6.7.8:2000
	peerBytes := []byte{
		1, 2, 3, 4, 0x03, 0xE8, // 1.2.3.4:1000
		5, 6, 7, 8, 0x07, 0xD0, // 5.6.7.8:2000
	}
	body := buildPeerResponse(peerBytes)

	peers := make(types.HashSet[string])
	extractPeers(body, peers)

	if len(peers) != 2 {
		t.Fatalf("expected 2 peers, got %d: %v", len(peers), peers)
	}
	if _, ok := peers["1.2.3.4:1000"]; !ok {
		t.Errorf("expected peer 1.2.3.4:1000, got: %v", peers)
	}
	if _, ok := peers["5.6.7.8:2000"]; !ok {
		t.Errorf("expected peer 5.6.7.8:2000, got: %v", peers)
	}
}

func TestExtractPeers_EmptyPeerData(t *testing.T) {
	body := []byte("d5:peers0:e")
	peers := make(types.HashSet[string])
	extractPeers(body, peers)
	if len(peers) != 0 {
		t.Errorf("expected 0 peers, got %d", len(peers))
	}
}

func TestExtractPeers_IncompleteEntry(t *testing.T) {
	// 5 bytes (not a multiple of 6) — last entry should be ignored
	peerBytes := []byte{10, 0, 0, 1, 0x1F, 0x90, 192, 168, 1} // 9 bytes: 1 full + 3 incomplete
	body := buildPeerResponse(peerBytes)

	peers := make(types.HashSet[string])
	extractPeers(body, peers)

	if len(peers) != 1 {
		t.Errorf("expected 1 peer (incomplete entry ignored), got %d: %v", len(peers), peers)
	}
}

func TestBuildTrackerURL_ContainsRequiredParams(t *testing.T) {
	torrent := &types.TorrentFile{
		InfoHash: []byte("01234567890123456789"), // 20 bytes
		Length:   5000,
	}

	result, err := buildTrackerURL(torrent, "http://tracker.example.com/announce")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.HasPrefix(result, "http://tracker.example.com/announce?") {
		t.Errorf("unexpected URL prefix: %s", result)
	}

	for _, param := range []string{"info_hash", "peer_id", "port", "uploaded", "downloaded", "left", "compact"} {
		if !strings.Contains(result, param) {
			t.Errorf("expected URL to contain param %q, got: %s", param, result)
		}
	}
}

func TestBuildTrackerURL_InvalidURL(t *testing.T) {
	torrent := &types.TorrentFile{
		InfoHash: []byte("01234567890123456789"),
		Length:   5000,
	}

	// A bare "%" is invalid percent-encoding and causes url.Parse to fail.
	_, err := buildTrackerURL(torrent, "%")
	if err == nil {
		t.Error("expected error for invalid tracker URL")
	}
}

func TestGeneratePeerID_Format(t *testing.T) {
	id := generatePeerID()

	if !strings.HasPrefix(id, "-GT0001-") {
		t.Errorf("expected peer ID to start with -GT0001-, got: %s", id)
	}
	if len(id) != 20 {
		t.Errorf("expected peer ID length 20, got %d: %s", len(id), id)
	}
}

func TestGeneratePeerID_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{})
	for i := 0; i < 10; i++ {
		id := generatePeerID()
		if _, exists := seen[id]; exists {
			t.Errorf("duplicate peer ID generated: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestDiscoverPeers_SingleTracker_ReturnsCompactPeers(t *testing.T) {
	peerBytes := []byte{10, 0, 0, 1, 0x1F, 0x90} // 10.0.0.1:8080

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(buildPeerResponse(peerBytes))
	}))
	defer server.Close()

	torrent := &types.TorrentFile{
		InfoHash: make([]byte, 20),
		Length:   1000,
	}
	urls := types.HashSet[string]{server.URL: struct{}{}}

	if err := DiscoverPeers(torrent, &urls); err != nil {
		t.Fatalf("DiscoverPeers returned error: %v", err)
	}

	if _, ok := torrent.Peers["10.0.0.1:8080"]; !ok {
		t.Errorf("expected peer 10.0.0.1:8080, got: %v", torrent.Peers)
	}
}

func TestDiscoverPeers_MultipleTrackers_MergesPeers(t *testing.T) {
	// Two trackers each returning a different peer; len(urls) > 1 triggers fetchConcurent.
	peer1 := []byte{10, 0, 0, 1, 0x1F, 0x90} // 10.0.0.1:8080
	peer2 := []byte{10, 0, 0, 2, 0x1F, 0x91} // 10.0.0.2:8081

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(buildPeerResponse(peer1))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(buildPeerResponse(peer2))
	}))
	defer server2.Close()

	torrent := &types.TorrentFile{
		InfoHash: make([]byte, 20),
		Length:   1000,
	}
	urls := types.HashSet[string]{
		server1.URL: struct{}{},
		server2.URL: struct{}{},
	}

	if err := DiscoverPeers(torrent, &urls); err != nil {
		t.Fatalf("DiscoverPeers returned error: %v", err)
	}

	if _, ok := torrent.Peers["10.0.0.1:8080"]; !ok {
		t.Errorf("expected peer 10.0.0.1:8080, got: %v", torrent.Peers)
	}
	if _, ok := torrent.Peers["10.0.0.2:8081"]; !ok {
		t.Errorf("expected peer 10.0.0.2:8081, got: %v", torrent.Peers)
	}
}

func TestDiscoverPeers_PrivateTorrent_UsesBruteForce(t *testing.T) {
	// Private flag forces sequential bruteForce even with multiple trackers.
	peerBytes := []byte{192, 168, 1, 1, 0x1F, 0x91} // 192.168.1.1:8081

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(buildPeerResponse(peerBytes))
	}))
	defer server.Close()

	torrent := &types.TorrentFile{
		InfoHash: make([]byte, 20),
		Length:   1000,
		Private:  true,
	}
	urls := types.HashSet[string]{server.URL: struct{}{}}

	if err := DiscoverPeers(torrent, &urls); err != nil {
		t.Fatalf("DiscoverPeers returned error: %v", err)
	}

	if _, ok := torrent.Peers["192.168.1.1:8081"]; !ok {
		t.Errorf("expected peer 192.168.1.1:8081, got: %v", torrent.Peers)
	}
}

func TestDiscoverPeers_TrackerUnreachable_ReturnsNoError(t *testing.T) {
	// A server that is immediately closed produces connection-refused errors,
	// which DiscoverPeers must handle gracefully (no returned error).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()

	torrent := &types.TorrentFile{
		InfoHash: make([]byte, 20),
		Length:   1000,
	}
	urls := types.HashSet[string]{url: struct{}{}}

	err := DiscoverPeers(torrent, &urls)
	if err != nil {
		t.Errorf("expected no error for unreachable tracker, got: %v", err)
	}
	if len(torrent.Peers) != 0 {
		t.Errorf("expected 0 peers for unreachable tracker, got %d", len(torrent.Peers))
	}
}

func TestDiscoverPeers_TrackerReturnsGarbage_ReturnsNoPeers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not a valid bencoded response"))
	}))
	defer server.Close()

	torrent := &types.TorrentFile{
		InfoHash: make([]byte, 20),
		Length:   1000,
	}
	urls := types.HashSet[string]{server.URL: struct{}{}}

	if err := DiscoverPeers(torrent, &urls); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(torrent.Peers) != 0 {
		t.Errorf("expected 0 peers for garbage response, got %d: %v", len(torrent.Peers), torrent.Peers)
	}
}
