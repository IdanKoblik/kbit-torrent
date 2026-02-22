package types

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func capturePrint(b BencodeValue, indent int) string {
	// Save original stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the print
	b.Print(indent)

	// Restore stdout and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old

	return buf.String()
}

func TestBencodePrint(t *testing.T) {
	data := BencodeDict{
		"name": BencodeString("torrent"),
		"size": BencodeInt(12345),
		"files": BencodeList{
			BencodeString("file1.txt"),
			BencodeString("file2.txt"),
		},
	}

	output := capturePrint(data, 0)

	expectedSubstrings := []string{
		`"name": "torrent"`,
		`"size": 12345`,
		`"files": [`,
		`"file1.txt"`,
		`"file2.txt"`,
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(output, substr) {
			t.Errorf("expected output to contain %q, got:\n%s", substr, output)
		}
	}
}
