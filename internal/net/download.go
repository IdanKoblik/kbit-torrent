package net

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"kbit/internal/logger"
	"kbit/pkg/types"
)

type pieceWork struct {
	index  int
	hash   []byte
	length int
}

type pieceResult struct {
	index int
	data  []byte
}

type workQueue struct {
	mu    sync.Mutex
	items []pieceWork
}

func (q *workQueue) pop() (pieceWork, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return pieceWork{}, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

func (q *workQueue) push(pw pieceWork) {
	q.mu.Lock()
	q.items = append(q.items, pw)
	q.mu.Unlock()
}

func (q *workQueue) len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func Download(t *types.TorrentFile) error {
	if len(t.Pieces) == 0 {
		return fmt.Errorf("torrent has no piece hashes; cannot download")
	}
	if t.PieceLength == 0 {
		return fmt.Errorf("torrent has no piece length; cannot download")
	}
	if len(t.Peers) == 0 {
		return fmt.Errorf("no peers discovered; cannot download")
	}

	fmt.Fprintf(os.Stderr, "Validating peers...\n")
	validAddrs := validatePeers(t)
	if len(validAddrs) == 0 {
		return fmt.Errorf("no reachable peers found")
	}
	fmt.Fprintf(os.Stderr, "%d reachable peer(s) found\n", len(validAddrs))

	fmt.Fprintf(os.Stderr, "Connecting and collecting piece availability...\n")
	pcs := collectBitfields(validAddrs, t)
	if len(pcs) == 0 {
		return fmt.Errorf("no peers provided piece availability")
	}
	fmt.Fprintf(os.Stderr, "%d peer(s) ready for downloading\n", len(pcs))

	queue := buildRarestFirstQueue(pcs, t)

	f, err := os.Create(t.Name)
	if err != nil {
		return fmt.Errorf("creating output file %q: %w", t.Name, err)
	}
	defer f.Close()

	if err := f.Truncate(t.Length); err != nil {
		return fmt.Errorf("pre-allocating output file: %w", err)
	}

	numPieces := len(t.Pieces)
	resultCh := make(chan pieceResult, numPieces)

	var wg sync.WaitGroup
	var alivePeers atomic.Int64
	var remaining atomic.Int64
	alivePeers.Store(int64(len(pcs)))
	remaining.Store(int64(numPieces))

	for _, pc := range pcs {
		wg.Add(1)
		go peerWorker(pc, queue, resultCh, &wg, &alivePeers, &remaining)
	}

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	completed := 0
	var downloaded int64

	for completed < numPieces {
		select {
		case res := <-resultCh:
			offset := int64(res.index) * t.PieceLength
			if _, err := f.WriteAt(res.data, offset); err != nil {
				return fmt.Errorf("writing piece %d: %w", res.index, err)
			}
			completed++
			downloaded += int64(len(res.data))
			printProgress(downloaded, t.Length, int(alivePeers.Load()))
		case <-doneCh:
			for len(resultCh) > 0 {
				res := <-resultCh
				offset := int64(res.index) * t.PieceLength
				if _, err := f.WriteAt(res.data, offset); err != nil {
					return fmt.Errorf("writing piece %d: %w", res.index, err)
				}
				completed++
				downloaded += int64(len(res.data))
				printProgress(downloaded, t.Length, 0)
			}
			if completed < numPieces {
				fmt.Fprintln(os.Stderr, "")
				return fmt.Errorf("download incomplete: %d/%d pieces received (all peers disconnected)", completed, numPieces)
			}
		}
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stdout, "Download complete: %s\n", t.Name)
	return nil
}

func peerWorker(
	pc *PeerConn,
	queue *workQueue,
	resultCh chan<- pieceResult,
	wg *sync.WaitGroup,
	alivePeers *atomic.Int64,
	remaining *atomic.Int64,
) {
	defer wg.Done()
	defer alivePeers.Add(-1)
	defer pc.Close()

	for {
		if remaining.Load() == 0 {
			return
		}

		pw, ok := queue.pop()
		if !ok {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		data, err := downloadPiece(pc, pw)
		if err != nil {
			logger.Log.Warn("piece download failed",
				slog.String("peer", pc.Addr),
				slog.Int("piece", pw.index),
				slog.String("error", err.Error()),
			)
			queue.push(pw)
			return
		}

		if err := checkPieceHash(data, pw); err != nil {
			logger.Log.Warn("piece hash mismatch",
				slog.String("peer", pc.Addr),
				slog.Int("piece", pw.index),
			)
			queue.push(pw)
			return
		}

		remaining.Add(-1)
		resultCh <- pieceResult{index: pw.index, data: data}
	}
}

func validatePeers(t *types.TorrentFile) []string {
	type result struct {
		addr string
		ok   bool
	}

	sem := make(chan struct{}, 20)
	resCh := make(chan result, len(t.Peers))
	var wg sync.WaitGroup

	for addr := range t.Peers {
		wg.Add(1)
		sem <- struct{}{}
		go func(addr string) {
			defer wg.Done()
			defer func() { <-sem }()

			conn, err := Handshake(addr, t.InfoHash)
			if err != nil {
				logger.Log.Warn("peer unreachable",
					slog.String("addr", addr),
					slog.String("error", err.Error()),
				)
				resCh <- result{addr: addr, ok: false}
				return
			}
			conn.Close()
			logger.Log.Info("peer reachable", slog.String("addr", addr))
			resCh <- result{addr: addr, ok: true}
		}(addr)
	}

	wg.Wait()
	close(resCh)

	var valid []string
	for r := range resCh {
		if r.ok {
			valid = append(valid, r.addr)
		}
	}
	return valid
}

func collectBitfields(addrs []string, t *types.TorrentFile) []*PeerConn {
	var mu sync.Mutex
	var pcs []*PeerConn
	var wg sync.WaitGroup

	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			conn, err := Handshake(addr, t.InfoHash)
			if err != nil {
				logger.Log.Warn("handshake failed for download connection",
					slog.String("addr", addr),
					slog.String("error", err.Error()),
				)
				return
			}

			pc := NewPeerConn(conn, addr)
			pc.SetDeadline(time.Now().Add(30 * time.Second))

			if err := pc.SendMsg(MsgInterested, nil); err != nil {
				logger.Log.Warn("failed to send Interested",
					slog.String("addr", addr),
					slog.String("error", err.Error()),
				)
				pc.Close()
				return
			}

			unchokedWithoutBitfield := false
			deadline := time.Now().Add(15 * time.Second)

		loop:
			for time.Now().Before(deadline) {
				pc.SetDeadline(time.Now().Add(10 * time.Second))
				msg, err := pc.ReadMsg()
				if err != nil {
					logger.Log.Warn("read error during bitfield exchange",
						slog.String("addr", addr),
						slog.String("error", err.Error()),
					)
					pc.Close()
					return
				}
				if msg == nil {
					// keep-alive
					continue
				}
				switch msg.ID {
				case MsgBitfield:
					pc.Bitfield = msg.Payload
					break loop
				case MsgUnchoke:
					unchokedWithoutBitfield = true
					// Some seeders skip the bitfield — keep waiting a bit.
				case MsgChoke:
					logger.Log.Warn("peer choked us during setup", slog.String("addr", addr))
					pc.Close()
					return
				}
			}

			if len(pc.Bitfield) == 0 && unchokedWithoutBitfield {
				numBytes := (len(t.Pieces) + 7) / 8
				pc.Bitfield = make([]byte, numBytes)
				for i := range pc.Bitfield {
					pc.Bitfield[i] = 0xFF
				}
			}

			if len(pc.Bitfield) == 0 {
				logger.Log.Warn("peer provided no bitfield", slog.String("addr", addr))
				pc.Close()
				return
			}

			available := 0
			for i := range len(t.Pieces) {
				if pc.HasPiece(i) {
					available++
				}
			}
			logger.Log.Info("peer connected",
				slog.String("addr", addr),
				slog.Int("pieces_available", available),
				slog.Int("total_pieces", len(t.Pieces)),
			)

			pc.SetDeadline(time.Now().Add(30 * time.Second))

			mu.Lock()
			pcs = append(pcs, pc)
			mu.Unlock()
		}(addr)
	}

	wg.Wait()
	return pcs
}

func buildRarestFirstQueue(pcs []*PeerConn, t *types.TorrentFile) *workQueue {
	numPieces := len(t.Pieces)

	avail := make([]int, numPieces)
	for _, pc := range pcs {
		for i := range numPieces {
			if pc.HasPiece(i) {
				avail[i]++
			}
		}
	}

	indices := make([]int, numPieces)
	for i := range indices {
		indices[i] = i
	}
	sort.SliceStable(indices, func(a, b int) bool {
		return avail[indices[a]] < avail[indices[b]]
	})

	items := make([]pieceWork, numPieces)
	for i, idx := range indices {
		items[i] = pieceWork{
			index:  idx,
			hash:   t.Pieces[idx],
			length: calcPieceLen(t, idx),
		}
	}
	return &workQueue{items: items}
}

func calcPieceLen(t *types.TorrentFile, i int) int {
	if i == len(t.Pieces)-1 {
		return int(t.Length - int64(i)*t.PieceLength)
	}
	return int(t.PieceLength)
}

func downloadPiece(pc *PeerConn, pw pieceWork) ([]byte, error) {
	pc.SetDeadline(time.Now().Add(30 * time.Second))

	buf := make([]byte, pw.length)
	downloaded := 0
	requested := 0
	backlog := 0
	maxBacklog := 5 // pipelined requests in flight

	for downloaded < pw.length {
		for backlog < maxBacklog && requested < pw.length {
			blockLen := min(BlockSize, pw.length-requested)
			payload := buildRequestPayload(pw.index, requested, blockLen)
			if err := pc.SendMsg(MsgRequest, payload); err != nil {
				return nil, fmt.Errorf("sending request: %w", err)
			}
			requested += blockLen
			backlog++
		}

		// Refresh deadline for each read.
		pc.SetDeadline(time.Now().Add(30 * time.Second))

		msg, err := pc.ReadMsg()
		if err != nil {
			return nil, fmt.Errorf("reading piece data: %w", err)
		}
		if msg == nil {
			// keep-alive
			continue
		}

		switch msg.ID {
		case MsgPiece:
			if len(msg.Payload) < 8 {
				return nil, fmt.Errorf("piece message too short")
			}
			gotIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
			gotBegin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
			if gotIndex != pw.index {
				return nil, fmt.Errorf("piece index mismatch: got %d, want %d", gotIndex, pw.index)
			}
			blockData := msg.Payload[8:]
			if gotBegin+len(blockData) > pw.length {
				return nil, fmt.Errorf("block overflows piece boundary")
			}
			copy(buf[gotBegin:], blockData)
			downloaded += len(blockData)
			backlog--

		case MsgChoke:
			return nil, fmt.Errorf("peer choked us mid-download")

		case MsgHave:
			// Ignore during download.

		default:
			// Ignore unknown messages.
		}
	}

	return buf, nil
}

func checkPieceHash(data []byte, pw pieceWork) error {
	h := sha1.Sum(data)
	if string(h[:]) != string(pw.hash) {
		return fmt.Errorf("piece %d hash mismatch", pw.index)
	}
	return nil
}

func printProgress(done, total int64, peers int) {
	if total == 0 {
		return
	}
	pct := float64(done) / float64(total)
	const width = 30
	filled := int(pct * float64(width))

	var bar string
	if filled >= width {
		bar = strings.Repeat("=", width)
	} else {
		bar = strings.Repeat("=", filled) + ">" + strings.Repeat(" ", width-filled-1)
	}

	fmt.Fprintf(os.Stderr, "\rDownloading: [%s] %.0f%% (%s / %s) | %d peers",
		bar, pct*100, formatBytes(done), formatBytes(total), peers)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
