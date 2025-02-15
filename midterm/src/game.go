package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	board *Board
	ai    *AI
}

func NewGame() *Game {
	return &Game{
		board: NewBoard(),
		ai:    NewAI(),
	}
}

func (g *Game) Update() error {
	if g.board.gameOver {
		os.Exit(0)
		return nil
	}

	// Handle mouse click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		row, col := g.board.ScreenToBoard(x, y)
		if g.board.MakeMove(row, col, Player) {
			if g.board.CheckWin(Player) {
				g.board.gameOver = true
			} else {
				// AI move
				startTime := time.Now()
				_, bestMove := g.ai.Minimax(g.board, 3, -9999, 9999, true)
				elapsedTime := time.Since(startTime).Seconds()
				fmt.Printf("AI move: %d, %d, Elapsed time: %f\n", bestMove.Row, bestMove.Col, elapsedTime)
				g.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer)
				if g.board.CheckWin(AIPlayer) {
					g.board.gameOver = true
				}
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.board.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 512, 512
}
