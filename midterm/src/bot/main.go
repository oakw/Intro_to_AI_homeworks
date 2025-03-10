package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	humanMode        = 1
	gomokuOnlineMode = 2
	productionMode   = 3
)

type CmdArgs struct {
	Mode         int
	StudentID    string
	ServerURL    string
	StreamToFile string
	AIStarts     bool
}

func parseArgs() CmdArgs {
	modePtr := flag.String("mode", "production", "Game mode: human, online, or production")
	idPtr := flag.String("id", "", "Student ID for production mode")
	serverPtr := flag.String("server", "http://0.0.0.0:55555", "Server base URL (required for production mode)")
	streamPtr := flag.String("stream", "", "Stream game state to file for external viewing")
	aiStartsPtr := flag.Bool("ai-starts", false, "AI makes the first move in human mode")

	flag.Parse()

	args := CmdArgs{
		Mode:         humanMode,
		ServerURL:    *serverPtr,
		StreamToFile: *streamPtr,
		AIStarts:     *aiStartsPtr,
	}

	switch *modePtr {
	case "online":
		args.Mode = gomokuOnlineMode
	case "production":
		args.Mode = productionMode
		args.StudentID = *idPtr

		// Validate required arguments for production mode
		if args.StudentID == "" {
			fmt.Println("Error: Student ID is required for production mode")
			fmt.Println("Usage: ./app -mode=production -id=RDB00001 -server=http://example.com [-stream=output.json]")
			os.Exit(1)
		}
		if args.ServerURL == "" {
			fmt.Println("Error: Server URL is required for production mode")
			fmt.Println("Usage: ./app -mode=production -id=RDB00001 -server=http://example.com [-stream=output.json]")
			os.Exit(1)
		}
	default:
		args.Mode = humanMode
	}

	return args
}

func main() {
	args := parseArgs()
	gameManager := NewGameManager()

	// Enable streaming if file was provided
	if args.StreamToFile != "" {
		log.Printf("Streaming game state to: %s\n", args.StreamToFile)
		gameManager.EnableStreaming(args.StreamToFile)
	}

	gameManager.mode = args.Mode

	switch args.Mode {
	case productionMode:
		// Run in production mode as in the midterm
		log.Printf("Starting in production mode with ID: %s, Server: %s\n",
			args.StudentID, args.ServerURL)
		RunProductionMode(args.StudentID, args.ServerURL, gameManager)

	case gomokuOnlineMode:
		log.Printf("Starting in online mode\n")
		RunGomokuOnlineMode(gameManager)

	case humanMode:
		if args.AIStarts {
			log.Printf("Starting in human mode (AI starts)\n")
		} else {
			log.Printf("Starting in human mode (Player starts)\n")
		}
		RunHumanMode(gameManager, args.AIStarts)
	}
}
