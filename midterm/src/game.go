package main

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	board              *Board
	ai                 *AI
	size               int
	mode               int
	winner             int
	gomokuOnlineClient *GomokuOnlineClient
}

func NewGame(mode int) *Game {
	size := 16

	if mode == gomokuOnlineMode {
		size = 19
	}

	return &Game{
		board: NewBoard(),
		ai:    NewAI(),
		size:  size,
		mode:  mode,
	}
}

func (g *Game) Update() error {
	// Handle mouse click
	if g.mode == gomokuOnlineMode {
		// Online mode
		g.handleGomokuOnlineMode()

	} else {
		// Playing against real human
		g.handleHumanModeCycle()
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
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.size * 32, g.size * 32
}

func (g *Game) handleHumanModeCycle() {
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
}



func (g *Game) handleGomokuOnlineMode() {
	if g.gomokuOnlineClient == nil {
		g.gomokuOnlineClient = initGomokuOnlineClient()
	} else if g.board.gameOver {
		return
	}

	if g.gomokuOnlineClient.gameState == -1 {
		g.gomokuOnlineClient.gameState = 0
	} else {
		g.gomokuOnlineClient.makeMoveAndObserve()
	}

	fmt.Println("Online move: ", g.gomokuOnlineClient.onlineX, g.gomokuOnlineClient.onlineY)
	if g.board.MakeMove(g.gomokuOnlineClient.onlineX, g.gomokuOnlineClient.onlineY, Player) {
		if g.gomokuOnlineClient.gameState > 0 {
			g.board.gameOver = true

			if g.board.CheckWin(Player) {
				g.winner = Player
			} else {
				g.winner = AIPlayer
			}

			return
		}

		startTime := time.Now()
		bestMove := g.ai.NextMove(g.board, 3, -9999, 9999, true)
		elapsedTime := time.Since(startTime).Seconds()
		fmt.Printf("AI made move: %d, %d, Elapsed time: %f\n", bestMove.Row, bestMove.Col, elapsedTime)
		g.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer)
		g.gomokuOnlineClient.aiX = bestMove.Row
		g.gomokuOnlineClient.aiY = bestMove.Col

	} else {
		// fmt.Println("Invalid move")
		// time.Sleep(100 * time.Second)
		// os.Exit(0)
	}
}
