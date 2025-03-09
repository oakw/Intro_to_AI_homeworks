package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

const (
	minRequestInterval = 50 * time.Millisecond // Minimum 50ms between requests
	gameOverDelay      = 3 * time.Second       // Time to display final board state
)

// ServerGameState represents the game state from the server
type ServerGameState struct {
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

func NewServerGameState(serverURL string) *ServerGameState {
	return &ServerGameState{
		lastRequestAt: time.Now().Add(-minRequestInterval),
		overloadCount: 0,
		serverURL:     serverURL,
	}
}

func (gs *ServerGameState) waitForRateLimit() {
	elapsed := time.Since(gs.lastRequestAt)
	if elapsed < minRequestInterval {
		time.Sleep(minRequestInterval - elapsed)
	}
	gs.lastRequestAt = time.Now()
}

func (gs *ServerGameState) makeRequest(url string) error {
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

func (gs *ServerGameState) StartGame(studentID string) error {
	url := fmt.Sprintf("%s/%s/start", gs.serverURL, studentID)
	err := gs.makeRequest(url)
	if err != nil {
		return err
	}

	gs.isBotBlack = gs.Color == "black"
	return nil
}

func (gs *ServerGameState) MakeMove(studentID string, x, y int) error {
	url := fmt.Sprintf("%s/%s/%d/%d/%d", gs.serverURL, studentID, gs.GameID, x, y)
	gs.moveCounter++
	return gs.makeRequest(url)
}

func (gs *ServerGameState) IsMyTurn() bool {
	if gs.isBotBlack {
		return gs.Turn == "black"
	}
	return gs.Turn == "white"
}

func (gs *ServerGameState) IsGameOver() bool {
	return gs.GameStatus != "ONGOING"
}

func (gs *ServerGameState) ShouldExit() bool {
	return gs.GameStatus == "LEAVE"
}

func (gs *ServerGameState) UpdateBoardFromServer(board *Board) {
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

func (gs *ServerGameState) GetWinner() int {
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
	gameState     *ServerGameState
	needsReset    bool
	winner        int
	lastMoves     map[string]bool // Track moves to avoid duplicates in board.moves
	gameOverTime  time.Time
	gameOverState bool
}

func NewProductionModeGame(studentID string, serverURL string) *ProductionModeGame {
	return &ProductionModeGame{
		board:      NewBoard(),
		ai:         NewAI(),
		studentID:  studentID,
		gameState:  NewServerGameState(serverURL),
		needsReset: true,
		lastMoves:  make(map[string]bool),
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
			return fmt.Errorf("game should exit")
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
		// Calculate time budget for this move
		// Assuming average game will have 256 moves (128 per player)
		maxMovesRemaining := 256 - len(g.board.moves)
		if maxMovesRemaining < 1 {
			maxMovesRemaining = 1
		}

		// Calculate time budget for this move (in seconds)
		moveTimeLimit := g.gameState.TimeRemaining / (float64(maxMovesRemaining) / 2)

		// Adjust depth based on available time
		searchDepth := 3 // Default depth
		if moveTimeLimit > 5.0 {
			searchDepth = 4 // Deeper search if we have time
		} else if moveTimeLimit < 1.0 {
			searchDepth = 2 // Quick search if we're short on time
		}

		// Set the maximum execution time for the minimax search
		g.ai.SetMaxExecutionTime(moveTimeLimit)

		fmt.Printf("Time remaining: %.2f seconds, moves remaining: ~%d, move time budget: %.2f seconds, depth: %d\n",
			g.gameState.TimeRemaining, maxMovesRemaining, moveTimeLimit, searchDepth)

		bestMove := g.ai.NextMove(g.board, searchDepth, math.Inf(-1), math.Inf(1), true)

		g.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer)               // Apply move locally
		err := g.gameState.MakeMove(g.studentID, bestMove.Col, bestMove.Row) // Apply move to server
		if err != nil {
			g.needsReset = true
			return nil
		}

		if g.gameState.ShouldExit() {
			return fmt.Errorf("game should exit")
		}

	} else {
		// check for opponent's move by polling the server
		// poll every time in headless mode
		err := g.gameState.makeRequest(fmt.Sprintf("%s/%s/%d",
			g.gameState.serverURL, g.studentID, g.gameState.GameID))

		if err != nil {
			g.needsReset = true
			return nil
		}

		if g.gameState.ShouldExit() {
			return fmt.Errorf("game should exit")
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

// RunHeadlessGameLoop runs the game continuously without a GUI
func (g *ProductionModeGame) RunHeadlessGameLoop() {
	for {
		err := g.Update()
		if err != nil {
			return
		}
	}
}

// RunProductionMode runs the production mode game with optional streaming
func RunProductionMode(studentID string, serverURL string, gameManager *GameManager) {
	game := NewProductionModeGame(studentID, serverURL)

	// Link game board with game manager for streaming
	if gameManager != nil && gameManager.streamEnabled {
		game.board = gameManager.board
	}

	for {
		err := game.Update()
		if err != nil {
			return
		}

		if gameManager != nil && gameManager.streamEnabled {
			gameManager.board = game.board
			gameManager.winner = game.winner

			if len(game.board.moves) > 0 {
				lastMove := game.board.moves[len(game.board.moves)-1]
				gameManager.lastMoveRow = lastMove.Row
				gameManager.lastMoveCol = lastMove.Col

				// Determine who made the move based on move count
				if game.gameState.moveCounter%2 == 1 {
					gameManager.lastMoveBy = AIPlayer
				} else {
					gameManager.lastMoveBy = Player
				}
			}

			if game.gameState.IsMyTurn() {
				gameManager.turn = AIPlayer
			} else {
				gameManager.turn = Player
			}

			gameManager.board.gameOver = game.board.gameOver
			gameManager.StreamState()
		}
	}
}
