package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	Size     = 16 // Board size (5x5)
	WinLen   = 5
	Cell     = 32 // Cell size in pixels
	Empty    = 0
	Player   = 1
	AIPlayer = 2
)

type Board struct {
	grid     [Size][Size]int
	gameOver bool
	moves    []Move
}

func NewBoard() *Board {
	return &Board{}
}

func (b *Board) ApplyMove(move Move, player int) {
	b.grid[move.Row][move.Col] = player
}

func (b *Board) Copy() *Board {
	newBoard := &Board{}
	newBoard.gameOver = b.gameOver
	for i := range b.grid {
		copy(newBoard.grid[i][:], b.grid[i][:])
	}
	return newBoard
}

func (b *Board) MakeMove(row, col, player int) bool {
	if row >= 0 && row < Size && col >= 0 && col < Size && b.grid[row][col] == Empty {
		b.grid[row][col] = player
		b.moves = append(b.moves, Move{Row: row, Col: col})
		return true
	}
	return false
}

func (b *Board) ScreenToBoard(x, y int) (int, int) {
	return y / Cell, x / Cell
}

func (b *Board) CheckWin(player int) bool {
	return CheckWinCondition(b.grid, player)
}

func (b *Board) Draw(screen *ebiten.Image) {
	// Draw grid
	for i := 0; i <= Size; i++ {
		ebitenutil.DrawLine(screen, float64(i*Cell), 0, float64(i*Cell), float64(Size*Cell), color.White)
		ebitenutil.DrawLine(screen, 0, float64(i*Cell), float64(Size*Cell), float64(i*Cell), color.White)
	}

	// Draw pieces
	for i := 0; i < Size; i++ {
		for j := 0; j < Size; j++ {
			if b.grid[i][j] == Player {
				ebitenutil.DrawCircle(screen, float64(j*Cell+Cell/2), float64(i*Cell+Cell/2), 10, color.RGBA{0, 255, 0, 255})
			} else if b.grid[i][j] == AIPlayer {
				ebitenutil.DrawCircle(screen, float64(j*Cell+Cell/2), float64(i*Cell+Cell/2), 10, color.RGBA{255, 0, 0, 255})
			}
		}
	}
}

func (b *Board) evaluate() int {
	threatScore, favorableScore := b.getThreatAndFavorScores()

	return favorableScore - threatScore
}

func (b *Board) evaluatePossibleThreatsAndFavors() (map[Move]int, map[Move]int) {
	possibleMoves := GenerateMoves(b, AIPlayer)

	threats := make(map[Move]int)
	favorableMoves := make(map[Move]int)

	for _, move := range possibleMoves {
		b.grid[move.Row][move.Col] = Player
		threats[move], favorableMoves[move] = b.getThreatAndFavorScores()
		b.grid[move.Row][move.Col] = Empty
	}

	return threats, favorableMoves
}

func (b *Board) getThreatAndFavorScores() (int, int) {
	threatScore := 0
	favorableScore := 0

	getValueScore := func(consecutiveCount int) int {
		if consecutiveCount == 2 {
			return oneInARow
		} else if consecutiveCount == 3 {
			return twoInARow
		} else if consecutiveCount == 4 {
			return threeInARow
		} else if consecutiveCount >= 5 {
			return fourInARow
		}

		return 0
	}

	getConsecutiveCounts := func(row int, col int, consecutivePlayerCount int, consecutiveAICount int) (int, int) {
		if b.grid[row][col] == Player {
			consecutivePlayerCount += 1
			consecutiveAICount = 0

		} else if b.grid[row][col] == AIPlayer {
			consecutiveAICount += 1
			consecutivePlayerCount = 0

		} else {
			consecutivePlayerCount = 0
			consecutiveAICount = 0
		}

		return consecutivePlayerCount, consecutiveAICount
	}

	consecutivePlayerCount := 0
	consecutiveAICount := 0
	for row := 0; row < Size; row++ {
		for col := 0; col < Size; col++ {
			consecutivePlayerCount, consecutiveAICount = getConsecutiveCounts(row, col, consecutivePlayerCount, consecutiveAICount)
			threatScore += getValueScore(consecutivePlayerCount)
			favorableScore += getValueScore(consecutiveAICount)
		}
	}

	consecutivePlayerCount = 0
	consecutiveAICount = 0
	for col := 0; col < Size; col++ {
		for row := 0; row < Size; row++ {
			consecutivePlayerCount, consecutiveAICount = getConsecutiveCounts(row, col, consecutivePlayerCount, consecutiveAICount)
			threatScore += getValueScore(consecutivePlayerCount)
			favorableScore += getValueScore(consecutiveAICount)
		}
	}

	// Check diagonal via diagonal traversal
	for i := 0; i < 2*Size-1; i++ {
		consecutiveAICount = 0
		consecutivePlayerCount = 0
		for j := 0; j <= i; j++ {
			row := i - j
			col := j
			if row >= 0 && row < Size && col >= 0 && col < Size {
				consecutivePlayerCount, consecutiveAICount = getConsecutiveCounts(row, col, consecutivePlayerCount, consecutiveAICount)
				threatScore += getValueScore(consecutivePlayerCount)
				favorableScore += getValueScore(consecutiveAICount)
			}
		}
	}

	// Check the other diagonal
	for i := 0; i < 2*Size-1; i++ {
		consecutiveAICount = 0
		consecutivePlayerCount = 0
		for j := 0; j <= i; j++ {
			row := i - j
			col := Size - 1 - j
			if row >= 0 && row < Size && col >= 0 && col < Size {
				consecutivePlayerCount, consecutiveAICount = getConsecutiveCounts(row, col, consecutivePlayerCount, consecutiveAICount)
				threatScore += getValueScore(consecutivePlayerCount)
				favorableScore += getValueScore(consecutiveAICount)
			}
		}
	}

	return threatScore, favorableScore
}
