package main

import(
	"os"
	"fmt"
	"log/slog"
	"kbit/internal/logger"
	"kbit/internal/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: kbit <command> [args]")
		os.Exit(1)
	}

	logger.Init(slog.LevelInfo)

	cmdName := os.Args[1]
	cmd, err := cmd.FindCommand(cmdName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := cmd.Run(os.Args[2:]); err != nil {
		logger.Log.Error("ERROR: ", "", err)
		os.Exit(1)
	}
}
