package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
)

const (
	Size   = 16 // Board size (5x5)
	WinLen = 5
	Cell   = 32 // Cell size in pixels
	Empty  = 0
	Player = 1
	AIPlayer     = 2
)

type Board struct {
	grid     [Size][Size]int
	gameOver bool
}

func NewBoard() *Board {
	return &Board{}
}

func (b *Board) MakeMove(row, col, player int) bool {
	if row >= 0 && row < Size && col >= 0 && col < Size && b.grid[row][col] == Empty {
		b.grid[row][col] = player
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
