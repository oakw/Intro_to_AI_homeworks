package main

import (
	"math/rand"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

const DEFAULT_MAX_MINIMAX_EXECUTION_SECONDS = 2

// TranspositionTable entry stores computed positions
type TTEntry struct {
	depth int
	score float64
	flag  int // 0: exact, 1: lower bound, 2: upper bound
	move  Move
}

const (
	EXACT      = 0
	LOWERBOUND = 1
	UPPERBOUND = 2
)

type AI struct {
	moveCount                     int
	lastMinimaxExecutionStartTime time.Time
	transpositionTable            map[string]TTEntry
	nodesExplored                 int
	maxExecutionTimeSeconds       float64
	offensiveBonus                float64            // Bonus for offensive opportunities
	patternScores                 map[string]float64 // Store pattern scores
}

func NewAI() *AI {
	ai := &AI{
		moveCount:               0,
		transpositionTable:      make(map[string]TTEntry),
		maxExecutionTimeSeconds: DEFAULT_MAX_MINIMAX_EXECUTION_SECONDS,
		offensiveBonus:          1.5, // Prioritize offensive moves 1.5x more than defensive
		patternScores:           make(map[string]float64),
	}

	// Initialize pattern scores for offense - higher values for critical patterns
	ai.patternScores["_XXXX_"] = 10000.0 // Open four (winning)
	ai.patternScores["XXXX_"] = 1000.0   // Closed four
	ai.patternScores["_XXXX"] = 1000.0   // Closed four
	ai.patternScores["_XXX_"] = 800.0    // Open three
	ai.patternScores["XX_XX"] = 800.0    // Split three
	ai.patternScores["_XX_X_"] = 300.0   // Potential three
	ai.patternScores["_X_XX_"] = 300.0   // Potential three
	ai.patternScores["__XX__"] = 200.0   // Open two

	return ai
}

// SetMaxExecutionTime allows setting a custom time limit for the minimax search
func (ai *AI) SetMaxExecutionTime(seconds float64) {
	if seconds <= 0 {
		seconds = DEFAULT_MAX_MINIMAX_EXECUTION_SECONDS
	}
	ai.maxExecutionTimeSeconds = seconds
}

func (ai *AI) NextMove(board *Board, depth int, alpha, beta float64, isMaximizing bool) Move {
	ai.moveCount += 1
	possibleMoves := GenerateMoves(board, AIPlayer)

	if ai.moveCount == 1 || len(board.moves) == 0 && len(possibleMoves) > 5 {
		// First move, play near the center
		return possibleMoves[rand.Intn(5 + 1)]
		
	} else if len(possibleMoves) < 1 {
		// No moves
		fmt.Println("No possible moves")
		os.Exit(0)
	}

	// Get both defensive threats and favorable opportunities integrated with AI's offensive patterns
	threats, favors, offensiveScores := ai.evaluateAllPatterns(board, possibleMoves)

	defensiveMoves := make([]Move, len(possibleMoves))
	winningMoves := make([]Move, len(possibleMoves))
	offensiveMoves := make([]Move, len(possibleMoves))

	copy(defensiveMoves, possibleMoves)
	copy(winningMoves, possibleMoves)
	copy(offensiveMoves, possibleMoves)

	fmt.Println("")

	sort.Slice(defensiveMoves, func(i, j int) bool {
		return threats[defensiveMoves[i]] > threats[defensiveMoves[j]]
	})

	sort.Slice(winningMoves, func(i, j int) bool {
		return favors[winningMoves[i]] > favors[winningMoves[j]]
	})

	sort.Slice(offensiveMoves, func(i, j int) bool {
		return offensiveScores[offensiveMoves[i]] > offensiveScores[offensiveMoves[j]]
	})

	if len(defensiveMoves) > 0 {
		fmt.Println("Top defensive move", defensiveMoves[0], "score", threats[defensiveMoves[0]])
	}
	if len(winningMoves) > 0 {
		fmt.Println("Top winning move", winningMoves[0], "score", favors[winningMoves[0]])
	}
	if len(offensiveMoves) > 0 {
		fmt.Println("Top offensive move", offensiveMoves[0], "score", offensiveScores[offensiveMoves[0]])
	}

	// 1. Look for immediate winning moves
	for _, move := range possibleMoves {
		tempBoard := board.Copy()
		tempBoard.grid[move.Row][move.Col] = AIPlayer
		if tempBoard.CheckWin(AIPlayer) {
			fmt.Println("Playing winning move", move)
			return move
		}
	}

	// 2. Block opponent's immediate winning moves
	for _, move := range possibleMoves {
		tempBoard := board.Copy()
		tempBoard.grid[move.Row][move.Col] = Player
		if tempBoard.CheckWin(Player) {
			fmt.Println("Blocking opponent's win", move)
			return move
		}
	}

	// 3. Create a four-in-a-row if available (high chance of winning next move)
	if len(winningMoves) > 0 && favors[winningMoves[0]] >= fourInARow {
		fmt.Println("Creating 4-in-a-row", winningMoves[0])
		return winningMoves[0]
	}

	// 4. Block opponent's four-in-a-row
	if len(defensiveMoves) > 0 && threats[defensiveMoves[0]] >= fourInARow {
		fmt.Println("Blocking opponent's 4-in-a-row", defensiveMoves[0])
		return defensiveMoves[0]
	}

	// 5. Create high-scoring offensive pattern (if score is significant)
	// This looks at our pattern-based offensive score which tends to be larger
	if len(offensiveMoves) > 0 && offensiveScores[offensiveMoves[0]] >= 800 {
		fmt.Println("Creating strong offensive pattern", offensiveMoves[0])
		return offensiveMoves[0]
	}

	// When heuristics are not immediate, do the minimax with increasing depth
	newBoard := board.Copy()
	ai.lastMinimaxExecutionStartTime = time.Now()
	ai.nodesExplored = 0
	ai.transpositionTable = make(map[string]TTEntry) // Clear TT for new search

	maxDepth := depth
	var bestMove Move = possibleMoves[0]
	var bestScore = offensiveScores[offensiveMoves[0]]

	for currentDepth := 2; currentDepth <= maxDepth; currentDepth++ {
		rootNode := MinimaxNode{board: *newBoard, lastMove: board.moves[len(board.moves)-1], nextMove: Move{-1, -1}, currentTurn: AIPlayer}
		score := ai.minimax(&rootNode, currentDepth, alpha, beta)

		// Stop if we're close to time limit
		if time.Since(ai.lastMinimaxExecutionStartTime).Seconds() > ai.maxExecutionTimeSeconds*0.9 {
			break
		}

		bestMove = rootNode.nextMove
		bestScore = int(score)
		fmt.Printf("Depth %d complete, best move: %v, score: %.2f, nodes: %d\n",
			currentDepth, bestMove, score, ai.nodesExplored)
	}

	if bestScore < threeInARow {
		fmt.Println("No good minmax moves found, playing best offensive move")
		return offensiveMoves[0]
	}

	fmt.Println("Minimax move", bestMove, "score", bestScore)

	return bestMove
}

// board evaluation and pattern recognition in a single pass
func (ai *AI) evaluateAllPatterns(board *Board, moves []Move) (map[Move]int, map[Move]int, map[Move]int) {
	threats := make(map[Move]int)
	favors := make(map[Move]int)
	offensiveScores := make(map[Move]int)

	for _, move := range moves {
		tempBoard := board.Copy()

		// Check defensive move (what if opponent plays here)
		tempBoard.grid[move.Row][move.Col] = Player
		threatScore, _ := tempBoard.getThreatAndFavorScores()
		threats[move] = threatScore
		tempBoard.grid[move.Row][move.Col] = Empty

		// Check favorable move (what if we play here)
		tempBoard.grid[move.Row][move.Col] = AIPlayer
		_, favorScore := tempBoard.getThreatAndFavorScores()
		favors[move] = favorScore

		// Now evaluate offensive patterns and augment with additional pattern recognition
		offensiveScore := 0

		// Check for patterns in all 8 directions
		directions := []struct{ dr, dc int }{
			{-1, 0}, {1, 0}, // vertical
			{0, -1}, {0, 1}, // horizontal
			{-1, -1}, {1, 1}, // diagonal \
			{-1, 1}, {1, -1}, // diagonal /
		}

		// Track if this move creates multiple threats (fork)
		threatDirections := 0

		// Check each direction for patterns
		for d := 0; d < 4; d++ {
			line := ""
			// Get cells in both directions for pattern matching
			for offset := -5; offset <= 5; offset++ {
				r := move.Row + directions[d*2].dr*offset
				c := move.Col + directions[d*2].dc*offset
				if r >= 0 && r < Size && c >= 0 && c < Size {
					switch tempBoard.grid[r][c] {
					case AIPlayer:
						line += "X"
					case Player:
						line += "O"
					case Empty:
						line += "_"
					}
				}
			}

			// Score patterns in this direction
			dirScore := 0
			hasPattern := false
			for pattern, score := range ai.patternScores {
				if containsPattern(line, pattern) {
					dirScore += int(score)
					hasPattern = true

					// If this is a critical pattern (open 3 or better), count it as a threat direction
					if score >= 800 {
						threatDirections++
					}
				}
			}

			if hasPattern {
				offensiveScore += dirScore
			}
		}

		// Special bonus for fork moves (creating threats in multiple directions)
		if threatDirections >= 2 {
			offensiveScore = int(float64(offensiveScore) * 2.0)
		}

		offensiveScores[move] = offensiveScore
		tempBoard.grid[move.Row][move.Col] = Empty
	}

	return threats, favors, offensiveScores
}

func containsPattern(s, pattern string) bool {
	for i := 0; i <= len(s)-len(pattern); i++ {
		matched := true
		for j := 0; j < len(pattern); j++ {
			if s[i+j] != pattern[j] {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

type MinimaxNode struct {
	board       Board
	lastMove    Move
	nextMove    Move
	currentTurn int
}

// Create a hash key for the board position for transposition table
func (node *MinimaxNode) hashKey() string {
	key := ""
	for i := 0; i < Size; i++ {
		for j := 0; j < Size; j++ {
			key += fmt.Sprintf("%d", node.board.grid[i][j]+1)
		}
	}
	key += fmt.Sprintf(":%d", node.currentTurn)
	return key
}

func (node *MinimaxNode) generateChildNodes() []MinimaxNode {
	moves := GenerateMoves(&node.board, node.currentTurn)
	childNodes := make([]MinimaxNode, len(moves))

	// Order moves by potential (helps alpha-beta pruning)
	moveScores := make(map[Move]int)
	for _, move := range moves {
		// Check if move is near existing stones (more likely to be good)
		nearStones := 0
		for dr := -2; dr <= 2; dr++ {
			for dc := -2; dc <= 2; dc++ {
				r, c := move.Row+dr, move.Col+dc
				if r >= 0 && r < Size && c >= 0 && c < Size &&
					node.board.grid[r][c] != Empty {
					nearStones++

					// Give extra weight to moves near own stones (offensive bias)
					if node.board.grid[r][c] == node.currentTurn {
						nearStones++
					}
				}
			}
		}
		moveScores[move] = nearStones
	}

	// Sort moves by score
	sort.Slice(moves, func(i, j int) bool {
		return moveScores[moves[i]] > moveScores[moves[j]]
	})

	for i, move := range moves {
		childNodes[i] = MinimaxNode{
			board:       node.board,
			lastMove:    move,
			nextMove:    Move{-1, -1},
			currentTurn: node.currentTurn * -1, // Switch to other player as per minimax
		}
		childNodes[i].board.grid[move.Row][move.Col] = node.currentTurn
	}

	return childNodes
}

func (ai *AI) minimax(node *MinimaxNode, depth int, alpha, beta float64) float64 {
	ai.nodesExplored++

	if node.board.CheckWin(AIPlayer) {
		return 10000.0 + float64(depth)
	}
	if node.board.CheckWin(Player) {
		return -10000.0 - float64(depth)
	}
	if depth == 0 || time.Since(ai.lastMinimaxExecutionStartTime).Seconds() > ai.maxExecutionTimeSeconds {
		threatScore, favorScore := node.board.getThreatAndFavorScores()
		score := favorScore - threatScore

		if node.currentTurn == AIPlayer {
			centerX, centerY := Size/2, Size/2
			for r := centerY - 2; r <= centerY+2; r++ {
				for c := centerX - 2; c <= centerX+2; c++ {
					if r >= 0 && r < Size && c >= 0 && c < Size && node.board.grid[r][c] == AIPlayer {
						// Small bonus for pieces controlling center
						score += 5
					}
				}
			}

			// Bonus for creating connections between existing stones
			if len(node.board.moves) > 3 {
				lastMove := node.lastMove
				nearbyStones := 0

				// Count own stones nearby the last move
				for dr := -2; dr <= 2; dr++ {
					for dc := -2; dc <= 2; dc++ {
						r, c := lastMove.Row+dr, lastMove.Col+dc
						if r >= 0 && r < Size && c >= 0 && c < Size &&
							node.board.grid[r][c] == AIPlayer && !(dr == 0 && dc == 0) {
							nearbyStones++
						}
					}
				}

				// Add bonus for connected stones
				score += nearbyStones * 10
			}
		}

		return float64(score)
	}

	// Check transposition table
	hashKey := node.hashKey()
	if entry, found := ai.transpositionTable[hashKey]; found && entry.depth >= depth {
		if entry.flag == EXACT {
			return entry.score
		} else if entry.flag == LOWERBOUND && entry.score > alpha {
			alpha = entry.score
		} else if entry.flag == UPPERBOUND && entry.score < beta {
			beta = entry.score
		}

		if alpha >= beta {
			return entry.score
		}
	}

	originalAlpha := alpha
	originalBeta := beta
	var bestMove Move
	var bestScore float64

	if node.currentTurn == AIPlayer {
		bestScore = -math.MaxFloat64
		for _, childNode := range node.generateChildNodes() {
			eval := ai.minimax(&childNode, depth-1, alpha, beta)
			if eval > bestScore {
				bestScore = eval
				bestMove = childNode.lastMove
				node.nextMove = bestMove
			}
			alpha = math.Max(alpha, eval)
			if beta <= alpha {
				break
			}
		}
	} else {
		bestScore = math.MaxFloat64
		for _, childNode := range node.generateChildNodes() {
			eval := ai.minimax(&childNode, depth-1, alpha, beta)
			if eval < bestScore {
				bestScore = eval
				bestMove = childNode.lastMove
				node.nextMove = bestMove
			}
			beta = math.Min(beta, eval)
			if beta <= alpha {
				break
			}
		}
	}

	// Store result in transposition table
	var flag int
	if bestScore <= originalAlpha {
		flag = UPPERBOUND
	} else if bestScore >= originalBeta {
		flag = LOWERBOUND
	} else {
		flag = EXACT
	}

	ai.transpositionTable[hashKey] = TTEntry{
		depth: depth,
		score: bestScore,
		flag:  flag,
		move:  bestMove,
	}

	return bestScore
}
