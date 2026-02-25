package types

type TorrentFile struct {
	Name        string
	InfoHash    []byte
	Length      int64
	PieceLength int64
	Pieces      [][]byte // each entry is a 20-byte SHA1 hash
	Private     bool

	TrackerURL string
	Trackers HashSet[string]
	Peers HashSet[string]
}
