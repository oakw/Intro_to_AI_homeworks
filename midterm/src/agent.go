package main

import (
	"math"
)

type AI struct{}

func NewAI() *AI {
	return &AI{}
}

func (ai *AI) Minimax(board *Board, depth int, alpha, beta float64, isMaximizing bool) (float64, Move) {
	if board.CheckWin(AIPlayer) {
		return 100, Move{}
	}
	if board.CheckWin(Player) {
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
				if board.grid[i][j] == Empty {
					board.grid[i][j] = AIPlayer
					eval, _ := ai.Minimax(board, depth-1, alpha, beta, false)
					board.grid[i][j] = Empty
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
				if board.grid[i][j] == Empty {
					board.grid[i][j] = Player
					eval, _ := ai.Minimax(board, depth-1, alpha, beta, true)
					board.grid[i][j] = Empty
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
