package types

type TorrentFile struct {
	Name string
	InfoHash []byte
	Length int64
	Private bool
}
