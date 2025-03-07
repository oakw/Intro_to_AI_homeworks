package main

import (
	"fmt"
	"time"
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
	// Handle based on game mode
	if g.mode == gomokuOnlineMode {
		// Online mode
		g.handleGomokuOnlineMode()
	} else {
		// Playing against real human - handled by external input
	}

	return nil
}

// ProcessPlayerMove handles a player move from external input
func (g *Game) ProcessPlayerMove(row, col int) bool {
	if g.board.MakeMove(row, col, Player) {
		if g.board.CheckWin(Player) {
			g.board.gameOver = true
			g.winner = Player
			return true
		}

		// AI response
		startTime := time.Now()
		bestMove := g.ai.NextMove(g.board, 3, -9999, 9999, true)
		elapsedTime := time.Since(startTime).Seconds()
		fmt.Printf("AI made move: %d, %d, Elapsed time: %f\n", bestMove.Row, bestMove.Col, elapsedTime)

		g.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer)
		if g.board.CheckWin(AIPlayer) {
			g.board.gameOver = true
			g.winner = AIPlayer
		}

		return true
	}
	return false
}

func (g *Game) GetBoardState() [Size][Size]int {
	return g.board.grid
}

func (g *Game) IsGameOver() bool {
	return g.board.gameOver
}

func (g *Game) GetWinner() int {
	return g.winner
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
	}
}
