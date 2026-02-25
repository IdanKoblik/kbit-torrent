package main

import(
	"os"
	"fmt"
	"log"
	"log/slog"
	"kbit/internal/logger"
	"kbit/internal/cmd"
	"kbit/internal/net"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: kbit <command> [args]")
		os.Exit(1)
	}

	verbose := false
	if len(os.Args) > 4 {
		if os.Args[4] == "verbose" {
			verbose = true
		}
	}

	logger.Init(slog.LevelInfo, verbose)
	id, err := net.GeneratePeerID()
	if err != nil {
		log.Fatalf("Error generating peer id: %v\n", err)
		return
	}

	net.PeerID = id

	cmdName := os.Args[1]
	cmd, err := cmd.FindCommand(cmdName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := cmd.Run(os.Args[2]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
