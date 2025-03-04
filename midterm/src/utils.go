package main

const (
	fourInARow  = 100 // Winning condition
	threeInARow = 50 
	twoInARow   = 10
	oneInARow   = 1
)

type Move struct {
	Row, Col int
}

func GenerateMoves(board *Board, player int) []Move {
	moves := []Move{}

	// From top-left corner
	// for i := 0; i < Size; i++ {
	// 	for j := 0; j < Size; j++ {
	// 		if board.grid[i][j] == Empty {
	// 			moves = append(moves, Move{i, j})
	// 		}
	// 	}
	// }

	// From center outwards
	center := Size / 2
	
	// Add center position first if empty
	if board.grid[center][center] == Empty {
		moves = append(moves, Move{center, center})
	}
	
	// For each radius from the center
	for offset := 1; offset <= center; offset++ {
		// Top and bottom rows of the current square
		for col := center - offset; col <= center + offset; col++ {
			// Top row
			row := center - offset
			if row >= 0 && col >= 0 && col < Size && board.grid[row][col] == Empty {
				moves = append(moves, Move{row, col})
			}
			
			// Bottom row
			row = center + offset
			if row < Size && col >= 0 && col < Size && board.grid[row][col] == Empty {
				moves = append(moves, Move{row, col})
			}
		}
		
		// Left and right columns of the current square (excluding corners already added)
		for row := center - offset + 1; row <= center + offset - 1; row++ {
			// Left column
			col := center - offset
			if row >= 0 && row < Size && col >= 0 && board.grid[row][col] == Empty {
				moves = append(moves, Move{row, col})
			}
			
			// Right column
			col = center + offset
			if row >= 0 && row < Size && col < Size && board.grid[row][col] == Empty {
				moves = append(moves, Move{row, col})
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
