package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	board  *Board
	ai     *AI
	winner int
}

func NewGame() *Game {
	return &Game{
		board: NewBoard(),
		ai:    NewAI(),
	}
}

func (g *Game) Update() error {
	// Handle mouse click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		row, col := g.board.ScreenToBoard(x, y)
		if g.board.MakeMove(row, col, Player) {
			if g.board.CheckWin(Player) {
				g.board.gameOver = true
				g.winner = Player
			} else {

				startTime := time.Now()
				bestMove := g.ai.NextMove(g.board, 3, -9999, 9999, true)
				elapsedTime := time.Since(startTime).Seconds()
				fmt.Printf("AI made move: %d, %d, Elapsed time: %f\n", bestMove.Row, bestMove.Col, elapsedTime)

				g.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer)
				if g.board.CheckWin(AIPlayer) {
					g.board.gameOver = true
					// Print the winner on the screen
					g.winner = AIPlayer
				}
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.board.Draw(screen)

	if g.board.gameOver {
		msg := "Player wins!"
		if g.winner == AIPlayer {
			msg = "AI wins!"
		}
		fmt.Println(msg)
		// wait until the user closes the window
		for !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {

		}

		os.Exit(0)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 512, 512
}
