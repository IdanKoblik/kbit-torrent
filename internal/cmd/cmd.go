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
	default:
		return nil, fmt.Errorf("unknown command: %s", name)
	}
}
