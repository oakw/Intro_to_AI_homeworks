package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	humanMode        = 1
	gomokuOnlineMode = 2
	productionMode   = 3
	// cmdMode = 4 // TODO:
)

type CmdArgs struct {
	Mode      int
	StudentID string
	Display   bool
	ServerURL string
}

func parseArgs() CmdArgs {
	modePtr := flag.String("mode", "human", "Game mode: human, online, or production")
	idPtr := flag.String("id", "", "Student ID for production mode")
	displayPtr := flag.Bool("display", false, "Enable display in production mode")
	serverPtr := flag.String("server", "", "Server base URL (required for production mode)")

	flag.Parse()

	args := CmdArgs{
		Mode:      humanMode,
		Display:   *displayPtr,
		ServerURL: *serverPtr,
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
			fmt.Println("Usage: ./app -mode=production -id=RDB00001 -server=http://example.com [-display]")
			os.Exit(1)
		}
		if args.ServerURL == "" {
			fmt.Println("Error: Server URL is required for production mode")
			fmt.Println("Usage: ./app -mode=production -id=RDB00001 -server=http://example.com [-display]")
			os.Exit(1)
		}
	default:
		args.Mode = humanMode
	}

	return args
}

func main() {
	args := parseArgs()

	if args.Mode == productionMode {
		// Run in production mode
		log.Printf("Starting in production mode with ID: %s (display: %v, server: %s)\n",
			args.StudentID, args.Display, args.ServerURL)
		RunProductionMode(args.StudentID, args.Display, args.ServerURL)
		return
	}

	game := NewGame(args.Mode)
	ebiten.SetWindowSize(512, 512)
	ebiten.SetWindowTitle("Gomoku")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
