package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	minRequestInterval = 50 * time.Millisecond // Minimum 50ms between requests
	gameOverDelay      = 3 * time.Second       // Time to display final board state
)

type GameState struct {
	RequestStatus string  `json:"request_status"`
	GameID        int     `json:"game_id"`
	GameStatus    string  `json:"game_status"`
	Color         string  `json:"color"`
	Turn          string  `json:"turn"`
	TimeRemaining float64 `json:"time_remaining"`
	Gameboard     [][]int `json:"gameboard"`
	lastRequestAt time.Time
	overloadCount int
	isBotBlack    bool
	moveCounter   int
	serverURL     string
}

func NewGameState(serverURL string) *GameState {
	return &GameState{
		lastRequestAt: time.Now().Add(-minRequestInterval),
		overloadCount: 0,
		serverURL:     serverURL,
	}
}

func (gs *GameState) waitForRateLimit() {
	elapsed := time.Since(gs.lastRequestAt)
	if elapsed < minRequestInterval {
		time.Sleep(minRequestInterval - elapsed)
	}
	gs.lastRequestAt = time.Now()
}

func (gs *GameState) makeRequest(url string) error {
	gs.waitForRateLimit()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if err := json.Unmarshal(body, gs); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	if gs.RequestStatus == "OVERLOAD" {
		// wait a bit and try again
		gs.overloadCount++
		time.Sleep(time.Duration(gs.overloadCount) * minRequestInterval)
		return gs.makeRequest(url)

	} else if gs.RequestStatus != "GOOD" {
		return fmt.Errorf("bad request status: %s", gs.RequestStatus)
	}

	return nil
}

func (gs *GameState) StartGame(studentID string) error {
	url := fmt.Sprintf("%s/%s/start", gs.serverURL, studentID)
	err := gs.makeRequest(url)
	if err != nil {
		return err
	}

	gs.isBotBlack = gs.Color == "black"
	return nil
}

func (gs *GameState) MakeMove(studentID string, x, y int) error {
	url := fmt.Sprintf("%s/%s/%d/%d/%d", gs.serverURL, studentID, gs.GameID, x, y)
	gs.moveCounter++
	return gs.makeRequest(url)
}

func (gs *GameState) IsMyTurn() bool {
	if gs.isBotBlack {
		return gs.Turn == "black"
	}
	return gs.Turn == "white"
}

func (gs *GameState) IsGameOver() bool {
	return gs.GameStatus != "ONGOING"
}

func (gs *GameState) ShouldExit() bool {
	return gs.GameStatus == "LEAVE"
}

func (gs *GameState) UpdateBoardFromServer(board *Board) {
	// Clear the board
	for i := 0; i < Size; i++ {
		for j := 0; j < Size; j++ {
			board.grid[i][j] = Empty
		}
	}

	// Map server values to local board
	for i := 0; i < len(gs.Gameboard); i++ {
		for j := 0; j < len(gs.Gameboard[i]); j++ {
			switch gs.Gameboard[i][j] {
			case 0: // Empty
				continue
			case 1: // Black
				if gs.isBotBlack {
					board.grid[i][j] = AIPlayer
				} else {
					board.grid[i][j] = Player
				}
			case 2: // White
				if gs.isBotBlack {
					board.grid[i][j] = Player
				} else {
					board.grid[i][j] = AIPlayer
				}
			}
		}
	}
}

func (gs *GameState) GetWinner() int {
	switch gs.GameStatus {
	case "BLACKWON":
		if gs.isBotBlack {
			return AIPlayer
		}
		return Player
	case "WHITEWON":
		if gs.isBotBlack {
			return Player
		}
		return AIPlayer
	default:
		return Empty
	}
}

type ProductionModeGame struct {
	board         *Board
	ai            *AI
	studentID     string
	gameState     *GameState
	needsReset    bool
	winner        int
	lastMoves     map[string]bool // Track moves to avoid duplicates in board.moves
	pollCounter   int
	gameOverTime  time.Time
	gameOverState bool
	headlessMode  bool
}

func NewProductionModeGame(studentID string, headless bool, serverURL string) *ProductionModeGame {
	return &ProductionModeGame{
		board:        NewBoard(),
		ai:           NewAI(),
		studentID:    studentID,
		gameState:    NewGameState(serverURL),
		needsReset:   true,
		lastMoves:    make(map[string]bool),
		headlessMode: headless,
	}
}

func (g *ProductionModeGame) Update() error {
	// wait a bit after game over to display final board state
	if g.gameOverState {
		if time.Since(g.gameOverTime) >= gameOverDelay {
			g.needsReset = true
			g.gameOverState = false
		}

		return nil
	}

	// Start new game if needed
	if g.needsReset {
		err := g.gameState.StartGame(g.studentID)
		if err != nil {
			time.Sleep(time.Second)
			return nil
		}

		if g.gameState.ShouldExit() {
			return ebiten.Termination
		}

		g.board = NewBoard()
		g.lastMoves = make(map[string]bool)
		g.winner = 0
		g.needsReset = false

		return nil
	}

	// Check if game is over, process the final state then
	if g.gameState.IsGameOver() {
		g.processCurrentBoardState()

		g.winner = g.gameState.GetWinner()
		g.board.gameOver = true
		g.gameOverState = true
		g.gameOverTime = time.Now()
		return nil
	}

	g.processCurrentBoardState()

	if g.gameState.IsMyTurn() {
		bestMove := g.ai.NextMove(g.board, 3, -9999, 9999, true)

		g.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer)               // Apply move locally
		err := g.gameState.MakeMove(g.studentID, bestMove.Col, bestMove.Row) // Apply move to server
		if err != nil {
			g.needsReset = true
			return nil
		}

		if g.gameState.ShouldExit() {
			return ebiten.Termination
		}

	} else {
		// check for opponent's move by polling the server

		if !g.headlessMode {
			// in GUI mode every 5th frame is polled
			g.pollCounter++
			if g.pollCounter < 5 {
				return nil
			}
			g.pollCounter = 0
		}

		err := g.gameState.makeRequest(fmt.Sprintf("%s/%s/%d/state",
			g.gameState.serverURL, g.studentID, g.gameState.GameID))

		if err != nil {
			g.needsReset = true
			return nil
		}

		if g.gameState.ShouldExit() {
			return ebiten.Termination
		}
	}

	return nil
}

func (g *ProductionModeGame) processCurrentBoardState() {
	// Generate a fresh board based on the server state
	tempBoard := [Size][Size]int{}

	for i := 0; i < len(g.gameState.Gameboard); i++ {
		for j := 0; j < len(g.gameState.Gameboard[i]); j++ {
			if g.gameState.Gameboard[i][j] == 0 {
				continue
			}

			var player int
			if (g.gameState.Gameboard[i][j] == 1 && !g.gameState.isBotBlack) ||
				(g.gameState.Gameboard[i][j] == 2 && g.gameState.isBotBlack) {
				player = Player
			} else {
				player = AIPlayer
			}

			// For some reason, move may not be tracked previously
			// Mark it also locally just to be safe
			moveKey := fmt.Sprintf("%d,%d", i, j)
			if !g.lastMoves[moveKey] {
				g.lastMoves[moveKey] = true
				g.board.MakeMove(i, j, player)
			}

			tempBoard[i][j] = player
		}
	}

	g.board.grid = tempBoard

	if g.gameState.IsGameOver() {
		g.board.gameOver = true
	}
}

func (g *ProductionModeGame) Draw(screen *ebiten.Image) {
	if !g.headlessMode {
		g.board.Draw(screen)
	}
}

func (g *ProductionModeGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return Size * Cell, Size * Cell
}

// RunHeadlessGameLoop runs the game continuously without a GUI
func (g *ProductionModeGame) RunHeadlessGameLoop() {
	for {
		err := g.Update()
		if err == ebiten.Termination {
			return
		}
	}
}

func RunProductionMode(studentID string, displayEnabled bool, serverURL string) {
	game := NewProductionModeGame(studentID, !displayEnabled, serverURL)

	if displayEnabled {
		ebiten.SetWindowSize(Size*Cell, Size*Cell)
		ebiten.SetWindowTitle(fmt.Sprintf("Gomoku - Production Mode (ID: %s, Server: %s)", studentID, serverURL))
		ebiten.RunGame(game)

	} else {
		game.RunHeadlessGameLoop()
	}
}
