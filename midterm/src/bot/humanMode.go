package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// RunHumanMode runs the game in human mode with viewer interaction
func RunHumanMode(gameManager *GameManager, aiStarts bool) {
	fmt.Println("Running in human mode")

	if gameManager.streamEnabled {
		fmt.Println("Use the viewer application to interact with the game")
	} else {
		fmt.Println("Enter moves as 'row col' (e.g. '7 8')")
	}

	gameManager.board = NewBoard()

	// Set initial turn based on who starts
	if aiStarts {
		gameManager.turn = AIPlayer
	} else {
		gameManager.turn = Player
	}

	gameManager.StreamState()

	var scanner *bufio.Scanner
	if !gameManager.streamEnabled {
		scanner = bufio.NewScanner(os.Stdin)
	}

	for {
		if gameManager.board.gameOver {
			gameManager.StreamState()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if gameManager.turn == Player {
			// Player's turn
			var moveReceived bool
			var row, col int

			if gameManager.streamEnabled {
				// Get move from viewer through file
				row, col, moveReceived = gameManager.ReadMoveFromViewer()
			} else {
				// Get move from console input as a pair of integers (row, col)
				if scanner != nil && scanner.Scan() {
					input := scanner.Text()
					moveReceived = true

					fields := strings.Fields(input)
					if len(fields) != 2 {
						fmt.Println("Invalid input. Please enter row and column as two integers.")
						moveReceived = false
					} else {
						var err error
						row, err = strconv.Atoi(fields[0])
						if err != nil {
							fmt.Println("Invalid row number:", err)
							moveReceived = false
						}

						col, err = strconv.Atoi(fields[1])
						if err != nil {
							fmt.Println("Invalid column number:", err)
							moveReceived = false
						}
					}
				}
			}

			if moveReceived {
				if gameManager.board.MakeMove(row, col, Player) {
					gameManager.RecordMove(row, col, Player)

					if gameManager.board.CheckWin(Player) {
						gameManager.SetGameOver(Player)
						if !gameManager.streamEnabled {
							fmt.Println("You win!")
						}
						continue
					}

					gameManager.UpdateTurn(AIPlayer)
				} else if !gameManager.streamEnabled {
					fmt.Println("Invalid move. Please try again.")
				}
			}
		} else {
			// AI's turn
			bestMove := gameManager.ai.NextMove(gameManager.board, 3, -9999, 9999, true)

			if gameManager.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer) {
				if !gameManager.streamEnabled {
					fmt.Printf("%d %d\n", bestMove.Row, bestMove.Col)
				}

				gameManager.RecordMove(bestMove.Row, bestMove.Col, AIPlayer)

				if gameManager.board.CheckWin(AIPlayer) {
					gameManager.SetGameOver(AIPlayer)
					continue
				}

				gameManager.UpdateTurn(Player)
			}
		}

		if gameManager.streamEnabled {
			time.Sleep(100 * time.Millisecond)
		}
	}
}
