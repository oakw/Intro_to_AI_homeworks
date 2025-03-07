package main

import (
	"fmt"
	"time"
)

// RunHumanMode runs the game in human mode with viewer interaction
func RunHumanMode(gameManager *GameManager) {
	fmt.Println("Running in human mode")
	fmt.Println("Use the viewer application to interact with the game")

	gameManager.turn = Player
	gameManager.board = NewBoard()
	gameManager.StreamState()

	for {
		if gameManager.board.gameOver {
			gameManager.StreamState()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if gameManager.turn == Player {
			// Player's turn - check for moves from viewer
			row, col, moveReceived := gameManager.ReadMoveFromViewer()
			if moveReceived {
				fmt.Printf("Received player move: %d,%d\n", row, col)
				if gameManager.board.MakeMove(row, col, Player) {
					gameManager.RecordMove(row, col, Player)

					if gameManager.board.CheckWin(Player) {
						gameManager.SetGameOver(Player)
						continue
					}

					// Switch turn to AI
					gameManager.UpdateTurn(AIPlayer)
				}
			}

		} else {
			// AI's turn
			bestMove := gameManager.ai.NextMove(gameManager.board, 3, -9999, 9999, true)

			if gameManager.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer) {
				fmt.Printf("AI made move: %d,%d\n", bestMove.Row, bestMove.Col)
				gameManager.RecordMove(bestMove.Row, bestMove.Col, AIPlayer)

				if gameManager.board.CheckWin(AIPlayer) {
					gameManager.SetGameOver(AIPlayer)
					continue
				}

				// Switch turn to player
				gameManager.UpdateTurn(Player)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}
