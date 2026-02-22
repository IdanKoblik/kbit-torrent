package types

type TorrentFile struct {
	InfoHash string
	Length string
	Private bool

	TrackerURL string
	Trackers map[string]struct{}
	Peers map[string]struct{}
}
