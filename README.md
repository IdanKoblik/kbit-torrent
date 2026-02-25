# kbit-torrent

A minimal BitTorrent CLI client written in Go.

> **IN DEV STAGE**

## Requirements

- Linux
- Go 1.25.6

## Installation

### Manual

```bash
git clone https://github.com/IdanKoblik/kbit-torrent.git
cd kbit-torrent
go build -o ./bin/kbit-torrent ./cmd/kbit-torrent
```

### Arch Linux

```bash
makepkg -si
```

## Usage

```
./bin/kbit-torrent <command> <file> [verbose]
```

## Commands

| Command     | Arguments | Description                                       |
|-------------|-----------|---------------------------------------------------|
| `parse`     | `<file>`  | Parse and display torrent metadata                |
| `handshake` | `<file>`  | Perform a BitTorrent handshake with a peer        |
| `download`  | `<file>`  | Download the torrent using rarest-first strategy  |

## Examples

Parse a torrent file:

```bash
./bin/kbit-torrent parse ./example.torrent
```

Perform a handshake with a peer (prompts for `host:port`):

```bash
./bin/kbit-torrent handshake ./example.torrent
```

Download a torrent:

```bash
./bin/kbit-torrent download ./example.torrent
```

Enable verbose logging by passing `verbose` as the fifth argument:

```bash
./bin/kbit-torrent parse ./example.torrent verbose
```

## Bugs

Report bugs at <https://github.com/IdanKoblik/kbit-torrent/issues>.
