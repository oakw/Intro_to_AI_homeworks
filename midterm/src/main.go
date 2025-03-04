package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	humanMode = 1
	gomokuOnlineMode = 2
	// productionMode = 3 // TODO:
	// cmdMode = 4 // TODO:
)

func main() {
	mode := humanMode
	if len(os.Args) > 1 && os.Args[1] == "online" {
		mode = gomokuOnlineMode
	}

	game := NewGame(mode)
	ebiten.SetWindowSize(512, 512)
	ebiten.SetWindowTitle("Gomoku")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
