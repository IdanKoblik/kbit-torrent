package cmd

import (
	"fmt"
)

type Command interface {
	Run(arg string) error
}

func FindCommand(name string) (Command, error) {
	switch name {
	case "parse":
		return &ParseCommand{}, nil
	case "handshake":
		return &HandshakeCommand{}, nil
	case "download":
		return &DownloadCommand{}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", name)
	}
}
