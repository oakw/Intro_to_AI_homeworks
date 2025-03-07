package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// GameState represents the serializable game state used for viewer communication
type GameState struct {
	Grid        [Size][Size]int `json:"grid"`
	GameOver    bool            `json:"gameOver"`
	Winner      int             `json:"winner"`
	LastMoveRow int             `json:"lastMoveRow"`
	LastMoveCol int             `json:"lastMoveCol"`
	LastMoveBy  int             `json:"lastMoveBy"`
	Size        int             `json:"size"`
	Turn        int             `json:"turn"`
	Mode        int             `json:"mode"`
	MoveRequest bool            `json:"moveRequest"`
	RequestRow  int             `json:"requestRow"`  // Row of requested move
	RequestCol  int             `json:"requestCol"`  // Column of requested move
}

// GameManager handles the game state and streaming to the viewer
type GameManager struct {
	board         *Board
	ai            *AI
	size          int
	mode          int
	winner        int
	lastMoveRow   int
	lastMoveCol   int
	lastMoveBy    int
	turn          int
	streamToFile  string
	streamEnabled bool
}

func NewGameManager() *GameManager {
	return &GameManager{
		board: NewBoard(),
		ai:    NewAI(),
		size:  Size,
		turn:  Player, // Player goes first by default
	}
}

func (gm *GameManager) EnableStreaming(filename string) {
	gm.streamToFile = filename
	gm.streamEnabled = true
}

func (gm *GameManager) GetState() GameState {
	return GameState{
		Grid:        gm.board.grid,
		GameOver:    gm.board.gameOver,
		Winner:      gm.winner,
		LastMoveRow: gm.lastMoveRow,
		LastMoveCol: gm.lastMoveCol,
		LastMoveBy:  gm.lastMoveBy,
		Size:        gm.size,
		Turn:        gm.turn,
		Mode:        gm.mode,
	}
}

// StreamState writes the current game state to a file for the viewer
func (gm *GameManager) StreamState() {
	if !gm.streamEnabled {
		return
	}

	state := gm.GetState()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Error serializing game state: %v", err)
		return
	}

	err = os.WriteFile(gm.streamToFile, data, 0644)
	if err != nil {
		log.Printf("Error writing game state to file: %v", err)
	} else {
		// Sync to disk to ensure file is written completely
		file, err := os.OpenFile(gm.streamToFile, os.O_RDONLY, 0)
		if err == nil {
			file.Sync()
			file.Close()
		}
	}
}

// RecordMove updates the game state with information about the last move
func (gm *GameManager) RecordMove(row, col int, player int) {
	gm.lastMoveRow = row
	gm.lastMoveCol = col
	gm.lastMoveBy = player

	// Check if this move results in a win
	if gm.board.CheckWin(player) {
		gm.board.gameOver = true
		gm.winner = player
	}

	// Stream the updated state to the viewer
	gm.StreamState()
}

func (gm *GameManager) UpdateTurn(nextPlayer int) {
	gm.turn = nextPlayer
	gm.StreamState()
}

func (gm *GameManager) SetGameOver(winner int) {
	gm.board.gameOver = true
	gm.winner = winner
	gm.StreamState()
}

// ReadMoveFromViewer reads move instructions from the viewer via the state file
func (gm *GameManager) ReadMoveFromViewer() (int, int, bool) {
	if !gm.streamEnabled {
		return 0, 0, false
	}

	data, err := os.ReadFile(gm.streamToFile)
	if err != nil {
		return 0, 0, false
	}

	if len(data) < 2 {
		return 0, 0, false
	}

	var state GameState
	err = json.Unmarshal(data, &state)
	if err != nil {
		fmt.Printf("Error parsing game state file: %v\n", err)
		return 0, 0, false
	}

	if state.MoveRequest {
		state.MoveRequest = false

		// Write the updated state back to the file
		updatedData, err := json.MarshalIndent(state, "", "  ")
		if err == nil {
			err = os.WriteFile(gm.streamToFile, updatedData, 0644)
			if err != nil {
				fmt.Printf("Error writing updated state: %v\n", err)
			}
		}

		return state.RequestRow, state.RequestCol, true
	}

	return 0, 0, false
}


func RunHeadlessHumanMode(gm *GameManager) {
	fmt.Println("Running in headless human mode")
	fmt.Println("Use the viewer application to interact with the game")

	gm.StreamState()

	// In headless mode, we just wait for input from the viewer
	for {
		time.Sleep(100 * time.Millisecond)

		// Check for player moves from viewer
		row, col, moveReceived := gm.ReadMoveFromViewer()
		if moveReceived && gm.board.MakeMove(row, col, Player) {
			gm.RecordMove(row, col, Player)

			if gm.board.gameOver {
				break
			}

			// AI's turn
			gm.UpdateTurn(AIPlayer)
			bestMove := gm.ai.NextMove(gm.board, 3, -9999, 9999, true)
			if gm.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer) {
				gm.RecordMove(bestMove.Row, bestMove.Col, AIPlayer)
				gm.UpdateTurn(Player)
			}
		}
	}
}
