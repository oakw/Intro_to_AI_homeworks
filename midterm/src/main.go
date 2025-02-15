package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	game := NewGame()
	ebiten.SetWindowSize(512, 512)
	ebiten.SetWindowTitle("Gomoku")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
