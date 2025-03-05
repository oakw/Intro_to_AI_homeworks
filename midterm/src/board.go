package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	Size     = 16 // TODO: Make dynamic
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

func getValueScore(consecutiveCount int, isOpen bool) int {
	if !isOpen && consecutiveCount < WinLen {
		// If the sequence is blocked and not already a win, reduce its value
		return getValueScore(consecutiveCount-1, true)
	}

	if consecutiveCount == 2 {
		return twoInARow
	} else if consecutiveCount == 3 {
		return threeInARow
	} else if consecutiveCount == 4 {
		return fourInARow
	} else if consecutiveCount >= 5 {
		return fiveInARow
	}

	return 0
}

func (b *Board) getThreatAndFavorScores() (int, int) {
	threatScore := 0
	favorableScore := 0

	// Check if a sequence can be potentially extended to reach WinLen
	isOpenSequence := func(row, col, dRow, dCol, player, count int) bool {
		// If already winning length, it doesn't matter if it's open
		if count >= WinLen {
			return true
		}

		// Check if there's space to extend before the sequence
		startRow, startCol := row-dRow, col-dCol
		hasSpaceBefore := startRow >= 0 && startRow < Size && startCol >= 0 && startCol < Size &&
			(b.grid[startRow][startCol] == Empty || b.grid[startRow][startCol] == player)

		// Check if there's space to extend after the sequence
		endRow, endCol := row+count*dRow, col+count*dCol
		hasSpaceAfter := endRow >= 0 && endRow < Size && endCol >= 0 && endCol < Size &&
			(b.grid[endRow][endCol] == Empty || b.grid[endRow][endCol] == player)

		// A sequence is open if it can be extended on at least one side
		return hasSpaceBefore || hasSpaceAfter
	}

	// Check horizontals (rows)
	for row := 0; row < Size; row++ {
		col := 0
		for col < Size {
			if b.grid[row][col] == Empty {
				col++
				continue
			}

			player := b.grid[row][col]
			startCol := col
			count := 0

			// Count consecutive pieces
			for col < Size && b.grid[row][col] == player {
				count++
				col++
			}

			// Check if this sequence can be extended
			isOpen := isOpenSequence(row, startCol, 0, 1, player, count)

			if player == Player {
				threatScore = max(threatScore, getValueScore(count, isOpen))
			} else if player == AIPlayer {
				favorableScore = max(favorableScore, getValueScore(count, isOpen))
			}
		}
	}

	// Check verticals (columns)
	for col := 0; col < Size; col++ {
		row := 0
		for row < Size {
			if b.grid[row][col] == Empty {
				row++
				continue
			}

			player := b.grid[row][col]
			startRow := row
			count := 0

			// Count consecutive pieces
			for row < Size && b.grid[row][col] == player {
				count++
				row++
			}

			// Check if this sequence can be extended
			isOpen := isOpenSequence(startRow, col, 1, 0, player, count)

			if player == Player {
				threatScore = max(threatScore, getValueScore(count, isOpen))
			} else if player == AIPlayer {
				favorableScore = max(favorableScore, getValueScore(count, isOpen))
			}
		}
	}

	// Check diagonals (top-left to bottom-right)
	for startRow := 0; startRow < Size; startRow++ {
		for startCol := 0; startCol < Size; startCol++ {
			row, col := startRow, startCol

			// Skip if out of bounds
			if row >= Size || col >= Size {
				continue
			}

			if b.grid[row][col] == Empty {
				continue
			}

			player := b.grid[row][col]
			count := 0

			// Count consecutive pieces
			for row < Size && col < Size && b.grid[row][col] == player {
				count++
				row++
				col++
			}

			// Only evaluate if we have at least 2 consecutive pieces
			if count >= 2 {
				isOpen := isOpenSequence(startRow, startCol, 1, 1, player, count)

				if player == Player {
					threatScore = max(threatScore, getValueScore(count, isOpen))
				} else if player == AIPlayer {
					favorableScore = max(favorableScore, getValueScore(count, isOpen))
				}
			}
		}
	}

	// Check diagonals (top-right to bottom-left)
	for startRow := 0; startRow < Size; startRow++ {
		for startCol := Size - 1; startCol >= 0; startCol-- {
			row, col := startRow, startCol

			// Skip if out of bounds
			if row >= Size || col < 0 {
				continue
			}

			if b.grid[row][col] == Empty {
				continue
			}

			player := b.grid[row][col]
			count := 0

			// Count consecutive pieces
			for row < Size && col >= 0 && b.grid[row][col] == player {
				count++
				row++
				col--
			}

			// Only evaluate if we have at least 2 consecutive pieces
			if count >= 2 {
				isOpen := isOpenSequence(startRow, startCol, 1, -1, player, count)

				if player == Player {
					threatScore = max(threatScore, getValueScore(count, isOpen))
				} else if player == AIPlayer {
					favorableScore = max(favorableScore, getValueScore(count, isOpen))
				}
			}
		}
	}

	return threatScore, favorableScore
}
