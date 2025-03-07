package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	Empty    = 0
	Player   = 1
	AIPlayer = 2
	Cell     = 32 // Cell size in pixels
)

// GameState matches the structure from the bot's interaction.go
type GameState struct {
	Grid        [16][16]int `json:"grid"`
	GameOver    bool        `json:"gameOver"`
	Winner      int         `json:"winner"`
	LastMoveRow int         `json:"lastMoveRow"`
	LastMoveCol int         `json:"lastMoveCol"`
	LastMoveBy  int         `json:"lastMoveBy"`
	Size        int         `json:"size"`
	Turn        int         `json:"turn"`
	Mode        int         `json:"mode"`
	MoveRequest bool        `json:"moveRequest"` // Added for requesting moves
	RequestRow  int         `json:"requestRow"`  // Row of requested move
	RequestCol  int         `json:"requestCol"`  // Column of requested move
}

type Viewer struct {
	stateFile     string
	gameState     GameState
	lastModified  time.Time
	size          int
	moveRequested bool
}

func NewViewer(stateFile string) *Viewer {
	return &Viewer{
		stateFile:     stateFile,
		size:          16, // Default size
		moveRequested: false,
	}
}

func (v *Viewer) Update() error {
	// Check if file has been modified
	info, err := os.Stat(v.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Wait for file to be created
			return nil
		}
		return nil // Don't propagate the error to prevent viewer from crashing
	}

	// Load state if file has been modified
	if info.ModTime().After(v.lastModified) {
		v.lastModified = info.ModTime()

		data, err := os.ReadFile(v.stateFile)
		if err != nil {
			return nil
		}

		// Check if file is empty or too small to be valid JSON
		if len(data) < 2 {
			return nil
		}

		err = json.Unmarshal(data, &v.gameState)
		if err != nil {
			fmt.Println("Error parsing game state:", err)
			return nil
		}

		if v.gameState.Size > 0 {
			v.size = v.gameState.Size
		}

		// Reset moveRequested if the bot acknowledged our move
		if !v.gameState.MoveRequest {
			v.moveRequested = false
		}
	}

	// Handle player input if it's the player's turn and game is not over
	if !v.gameState.GameOver && v.gameState.Turn == Player && !v.moveRequested {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			col, row := x/Cell, y/Cell

			// Validate move is within board bounds and on an empty cell
			if row >= 0 && row < v.size && col >= 0 && col < v.size && v.gameState.Grid[row][col] == Empty {
				// Send move to bot by updating the state file
				v.gameState.MoveRequest = true
				v.gameState.RequestRow = row
				v.gameState.RequestCol = col
				v.moveRequested = true

				data, err := json.MarshalIndent(v.gameState, "", "  ")
				if err != nil {
					fmt.Println("Error marshaling move:", err)
					return nil
				}

				err = os.WriteFile(v.stateFile, data, 0644)
				if err != nil {
					fmt.Println("Error writing move to file:", err)
					v.moveRequested = false
					return nil
				}

				fmt.Printf("Player move at row=%d, col=%d sent to bot\n", row, col)
			}
		}
	}

	return nil
}

func (v *Viewer) Draw(screen *ebiten.Image) {
	// Draw grid
	size := v.gameState.Size
	if size == 0 {
		size = 16 // Default size
	}

	// Fill background
	screen.Fill(color.RGBA{40, 40, 40, 255})

	// Draw grid lines
	for i := 0; i <= size; i++ {
		ebitenutil.DrawLine(screen, float64(i*Cell), 0, float64(i*Cell), float64(size*Cell), color.White)
		ebitenutil.DrawLine(screen, 0, float64(i*Cell), float64(size*Cell), float64(i*Cell), color.White)
	}

	// Draw pieces
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if v.gameState.Grid[i][j] == Player {
				ebitenutil.DrawCircle(screen, float64(j*Cell+Cell/2), float64(i*Cell+Cell/2), 10, color.RGBA{0, 255, 0, 255})
			} else if v.gameState.Grid[i][j] == AIPlayer {
				ebitenutil.DrawCircle(screen, float64(j*Cell+Cell/2), float64(i*Cell+Cell/2), 10, color.RGBA{255, 0, 0, 255})
			}
		}
	}

	// Highlight last move
	if v.gameState.LastMoveRow >= 0 && v.gameState.LastMoveCol >= 0 {
		ebitenutil.DrawRect(
			screen,
			float64(v.gameState.LastMoveCol*Cell),
			float64(v.gameState.LastMoveRow*Cell),
			Cell, Cell,
			color.RGBA{255, 255, 0, 64},
		)
	}

	// Display game status
	if v.gameState.GameOver {
		msg := "Game Over: "
		if v.gameState.Winner == Player {
			msg += "Player wins!"
		} else if v.gameState.Winner == AIPlayer {
			msg += "AI wins!"
		} else {
			msg += "Draw!"
		}
		ebitenutil.DebugPrintAt(screen, msg, 10, 10)
	} else {
		msg := "Current turn: "
		if v.gameState.Turn == Player {
			if v.moveRequested {
				msg += "Player (Move requested...)"
			} else {
				msg += "Player (Click to place a stone)"
			}
		} else {
			msg += "AI thinking..."
		}
		ebitenutil.DebugPrintAt(screen, msg, 10, 10)
	}
}

func (v *Viewer) Layout(outsideWidth, outsideHeight int) (int, int) {
	size := v.gameState.Size
	if size == 0 {
		size = 16 // Default size
	}
	return size * Cell, size * Cell
}

func main() {
	stateFilePtr := flag.String("state", "game_state.json", "Path to the game state file")
	flag.Parse()

	viewer := NewViewer(*stateFilePtr)

	ebiten.SetWindowSize(512, 512)
	ebiten.SetWindowTitle("Gomoku Viewer")

	fmt.Printf("Viewing game state from: %s\n", *stateFilePtr)
	fmt.Println("Click on the board to make a move when it's your turn")

	if err := ebiten.RunGame(viewer); err != nil {
		log.Fatal(err)
	}
}
