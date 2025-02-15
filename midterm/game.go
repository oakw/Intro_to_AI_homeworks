package main

import (
	"fmt"
	"math"
)

const (
	Size   = 16 // Board size (5x5 for simplicity)
	Empty  = 0 // Empty cell
	Player = 1 // Human player
	AI     = 2 // AI player
	WinLen = 5 // Number of pieces in a row to win
)

type Move struct {
	Row, Col int
}

var board [Size][Size]int

func printBoard() {
	fmt.Println("\nBoard:")
	for i := 0; i < Size; i++ {
		for j := 0; j < Size; j++ {
			switch board[i][j] {
			case Empty:
				fmt.Print(". ")
			case Player:
				fmt.Print("O ")
			case AI:
				fmt.Print("X ")
			}
		}
		fmt.Println()
	}
}

func isValidMove(row, col int) bool {
	return row >= 0 && row < Size && col >= 0 && col < Size && board[row][col] == Empty
}

func checkWin(player int) bool {
	// Check rows, columns, and diagonals for a win
	for i := 0; i < Size; i++ {
		for j := 0; j < Size; j++ {
			if j <= Size-WinLen && checkDirection(i, j, 0, 1, player) ||
				i <= Size-WinLen && checkDirection(i, j, 1, 0, player) ||
				(i <= Size-WinLen && j <= Size-WinLen && checkDirection(i, j, 1, 1, player)) ||
				(i >= WinLen-1 && j <= Size-WinLen && checkDirection(i, j, -1, 1, player)) {
				return true
			}
		}
	}
	return false
}

func checkDirection(row, col, dRow, dCol, player int) bool {
	for k := 0; k < WinLen; k++ {
		if board[row+k*dRow][col+k*dCol] != player {
			return false
		}
	}
	return true
}

func minimax(depth int, alpha, beta float64, isMaximizing bool) (float64, Move) {
	if checkWin(AI) {
		return 100, Move{}
	}
	if checkWin(Player) {
		return -100, Move{}
	}
	if depth == 0 {
		return 0, Move{}
	}

	bestMove := Move{-1, -1}
	if isMaximizing {
		maxEval := -math.MaxFloat64
		for i := 0; i < Size; i++ {
			for j := 0; j < Size; j++ {
				if board[i][j] == Empty {
					board[i][j] = AI
					eval, _ := minimax(depth-1, alpha, beta, false)
					board[i][j] = Empty
					if eval > maxEval {
						maxEval = eval
						bestMove = Move{i, j}
					}
					alpha = math.Max(alpha, eval)
					if beta <= alpha {
						break
					}
				}
			}
		}
		return maxEval, bestMove
	} else {
		minEval := math.MaxFloat64
		for i := 0; i < Size; i++ {
			for j := 0; j < Size; j++ {
				if board[i][j] == Empty {
					board[i][j] = Player
					eval, _ := minimax(depth-1, alpha, beta, true)
					board[i][j] = Empty
					if eval < minEval {
						minEval = eval
						bestMove = Move{i, j}
					}
					beta = math.Min(beta, eval)
					if beta <= alpha {
						break
					}
				}
			}
		}
		return minEval, bestMove
	}
}

func main() {
	fmt.Println("Welcome to Gomoku! You are 'O' and AI is 'X'.")

	for {
		printBoard()
		var row, col int
		fmt.Print("Enter your move (row and column): ")
		fmt.Scan(&row, &col)

		if !isValidMove(row, col) {
			fmt.Println("Invalid move! Try again.")
			continue
		}
		board[row][col] = Player

		if checkWin(Player) {
			printBoard()
			fmt.Println("You win!")
			break
		}

		_, bestMove := minimax(3, -math.MaxFloat64, math.MaxFloat64, true)
		if bestMove.Row == -1 {
			fmt.Println("Game Over! It's a draw.")
			break
		}
		board[bestMove.Row][bestMove.Col] = AI

		if checkWin(AI) {
			printBoard()
			fmt.Println("AI wins!")
			break
		}
	}
}
