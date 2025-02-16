package main

const (
	fourInARow  = 100 // Winning condition
	threeInARow = 50  // Winning condition
	twoInARow   = 10
	oneInARow   = 1
	// centerBonus = 3 // Bonus for controlling center
)

type Move struct {
	Row, Col int
}

func GenerateMoves(board *Board, player int) []Move {
	moves := []Move{}
	for i := 0; i < Size; i++ {
		for j := 0; j < Size; j++ {
			if board.grid[i][j] == Empty {
				moves = append(moves, Move{i, j})
			}
		}
	}
	return moves
}

func CheckWinCondition(grid [Size][Size]int, player int) bool {
	for i := 0; i < Size; i++ {
		for j := 0; j < Size; j++ {
			if j <= Size-WinLen && checkDirection(grid, i, j, 0, 1, player) ||
				i <= Size-WinLen && checkDirection(grid, i, j, 1, 0, player) ||
				(i <= Size-WinLen && j <= Size-WinLen && checkDirection(grid, i, j, 1, 1, player)) ||
				(i >= WinLen-1 && j <= Size-WinLen && checkDirection(grid, i, j, -1, 1, player)) {
				return true
			}
		}
	}
	return false
}

func checkDirection(grid [Size][Size]int, row, col, dRow, dCol, player int) bool {
	for k := 0; k < WinLen; k++ {
		if grid[row+k*dRow][col+k*dCol] != player {
			return false
		}
	}
	return true
}
