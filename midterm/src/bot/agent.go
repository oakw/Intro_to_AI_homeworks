package main

import (
	"fmt"
	"math"
	"math/rand"
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
	offensiveBonus                float64        // Bonus for offensive opportunities
	patternScores                 map[string]int
}

func NewAI() *AI {
	ai := &AI{
		moveCount:               0,
		transpositionTable:      make(map[string]TTEntry),
		maxExecutionTimeSeconds: DEFAULT_MAX_MINIMAX_EXECUTION_SECONDS,
		offensiveBonus:          1.5, // Prioritize offensive moves 1.5x more than defensive
		patternScores:           make(map[string]int),
	}

	// Initialize unified pattern scores - used for both offensive pattern recognition and board evaluation
	ai.patternScores["XXXXX"] = 100000 // Five in a row (win)
	ai.patternScores["_XXXX_"] = 10000 // Open four (virtually winning)
	ai.patternScores["XXXX_"] = 5000   // Half-open four
	ai.patternScores["_XXXX"] = 5000   // Half-open four
	ai.patternScores["_XXX_"] = 2000   // Open three
	ai.patternScores["XX_XX"] = 1500   // Split three
	ai.patternScores["XXX_"] = 500     // Half-open three
	ai.patternScores["_XXX"] = 500     // Half-open three
	ai.patternScores["XX_X"] = 400     // Broken three
	ai.patternScores["X_XX"] = 400     // Broken three
	ai.patternScores["_XX_X_"] = 300   // Potential three
	ai.patternScores["_X_XX_"] = 300   // Potential three
	ai.patternScores["_XX_"] = 200     // Open two
	ai.patternScores["XX_"] = 50       // Half-open two
	ai.patternScores["_XX"] = 50       // Half-open two
	ai.patternScores["X_X"] = 40       // Broken two
	ai.patternScores["__XX__"] = 200   // Open two
	ai.patternScores["__X__"] = 10     // Single stone with space

	return ai
}

// SetMaxExecutionTime allows setting a custom time limit for the minimax search
func (ai *AI) SetMaxExecutionTime(seconds float64) {
	if seconds <= 0 {
		seconds = DEFAULT_MAX_MINIMAX_EXECUTION_SECONDS
	}
	ai.maxExecutionTimeSeconds = seconds
}

func (ai *AI) NextMove(board *Board, depth int) Move {
	ai.moveCount += 1
	possibleMoves := GenerateMoves(board, AIPlayer)

	if ai.moveCount == 1 || len(board.moves) == 0 && len(possibleMoves) > 5 {
		// First move, play near the center
		return possibleMoves[rand.Intn(5+1)]

	} else if len(possibleMoves) < 1 {
		// No moves
		fmt.Println("No possible moves")
		os.Exit(0)
	}

	// Enhanced threat detection - immediate win detection
	// Check if opponent has any winning moves and if we can create a winning move
	immediateWinMove, immediateBlockMove := ai.detectImminentThreats(board, possibleMoves)
	
	if immediateWinMove.Row != -1 {
		fmt.Println("Playing immediate winning move", immediateWinMove)
		return immediateWinMove
	}
	
	if immediateBlockMove.Row != -1 {
		fmt.Println("Blocking opponent's immediate win", immediateBlockMove)
		return immediateBlockMove
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

	// Create a four-in-a-row if available (high chance of winning next move)
	if len(winningMoves) > 0 && favors[winningMoves[0]] >= fourInARow {
		fmt.Println("Creating 4-in-a-row", winningMoves[0])
		return winningMoves[0]
	}

	// Block opponent's four-in-a-row - this is critical for defense
	// Use a secondary check to ensure we don't miss any threats
	fourInARowBlocks := ai.detectFourInARowThreats(board, possibleMoves)
	if len(fourInARowBlocks) > 0 {
		fmt.Println("Blocking opponent's 4-in-a-row (enhanced detection)", fourInARowBlocks[0])
		return fourInARowBlocks[0]
	}
	
	// Standard 4-in-a-row detection as backup
	if len(defensiveMoves) > 0 && threats[defensiveMoves[0]] >= fourInARow {
		fmt.Println("Blocking opponent's 4-in-a-row", defensiveMoves[0])
		return defensiveMoves[0]
	}

	// When heuristics are not immediate, do the minimax with increasing depth
	newBoard := board.Copy()
	ai.lastMinimaxExecutionStartTime = time.Now()
	ai.nodesExplored = 0
	ai.transpositionTable = make(map[string]TTEntry) // Clear TT for new search

	maxDepth := depth
	var bestMove Move = offensiveMoves[0]
	var bestScore = offensiveScores[offensiveMoves[0]]

	for currentDepth := 2; currentDepth <= maxDepth; currentDepth++ {
		rootNode := MinimaxNode{board: *newBoard, lastMove: board.moves[len(board.moves)-1], nextMove: Move{-1, -1}, currentTurn: AIPlayer}
		score := ai.minimax(&rootNode, currentDepth, math.Inf(-1), math.Inf(1))

		// Stop if we're close to time limit
		if time.Since(ai.lastMinimaxExecutionStartTime).Seconds() > ai.maxExecutionTimeSeconds {
			break
		}

		bestMove = rootNode.nextMove
		bestScore = int(score)
		fmt.Printf("Depth %d complete, best move: %v, score: %.2f, nodes: %d\n",
			currentDepth, bestMove, score, ai.nodesExplored)
	}

	fmt.Println("Minimax move", bestMove, "score", bestScore)

	return bestMove
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
		return 100000.0 + float64(depth)
	}
	if node.board.CheckWin(Player) {
		return -100000.0 - float64(depth)
	}
	if depth == 0 || time.Since(ai.lastMinimaxExecutionStartTime).Seconds() > ai.maxExecutionTimeSeconds {
		return float64(ai.evaluateBoard(&node.board))
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

// comprehensive evaluation of the provided current board position
func (ai *AI) evaluateBoard(board *Board) int {
	score := 0

	aiPatterns := make(map[string]int)
	playerPatterns := make(map[string]int)

	// 1. Process all horizontal lines
	for row := 0; row < Size; row++ {
		ai.processLine(board, 0, row, 1, 0, aiPatterns, playerPatterns)
	}

	// 2. Process all vertical lines
	for col := 0; col < Size; col++ {
		ai.processLine(board, col, 0, 0, 1, aiPatterns, playerPatterns)
	}

	// 3. Process all diagonal lines (top-left to bottom-right)
	for col := 0; col < Size; col++ {
		ai.processLine(board, col, 0, 1, 1, aiPatterns, playerPatterns)
	}

	for row := 1; row < Size; row++ {
		ai.processLine(board, 0, row, 1, 1, aiPatterns, playerPatterns)
	}

	// 4. Process all diagonal lines (top-right to bottom-left)
	for col := 0; col < Size; col++ {
		ai.processLine(board, col, 0, -1, 1, aiPatterns, playerPatterns)
	}

	for row := 1; row < Size; row++ {
		ai.processLine(board, Size-1, row, -1, 1, aiPatterns, playerPatterns)
	}

	// Calculate score based on pattern counts
	for pattern, weight := range ai.patternScores {
		aiCount := aiPatterns[pattern]
		playerCount := playerPatterns[pattern]

		// Apply weights - offensive patterns have their listed weights
		// Defensive patterns (opponent threats) have higher weights to prioritize defense
		score += weight * aiCount
		score -= (weight + (weight / 5)) * playerCount // 20% higher weight for defense
	}

	// Add positional bonuses
	score += evaluatePositionalFactors(board)

	return score
}

// detectImminentThreats does a direct scan for immediate win/loss situations
func (ai *AI) detectImminentThreats(board *Board, moves []Move) (Move, Move) {
	var winningMove Move = Move{-1, -1}
	var blockingMove Move = Move{-1, -1}
	
	// Check for immediate wins
	for _, move := range moves {
		// Check if we can win with this move
		tempBoard := board.Copy()
		tempBoard.grid[move.Row][move.Col] = AIPlayer
		if tempBoard.CheckWin(AIPlayer) {
			winningMove = move
			break
		}
		
		// Check if opponent would win if they played here
		tempBoard = board.Copy()
		tempBoard.grid[move.Row][move.Col] = Player
		if tempBoard.CheckWin(Player) {
			blockingMove = move
			// Don't break - continue checking if there's a winning move
		}
	}
	
	return winningMove, blockingMove
}

// detectFourInARowThreats detects if opponent can create a 4-in-a-row that would lead to a win
func (ai *AI) detectFourInARowThreats(board *Board, moves []Move) []Move {
	threats := []Move{}
	
	for _, move := range moves {
		tempBoard := board.Copy()
		tempBoard.grid[move.Row][move.Col] = Player // Simulate opponent's move
		
		// Check if this creates a 4-in-a-row for opponent
		if ai.hasFourInARow(tempBoard, Player) {
			threats = append(threats, move)
		}
	}
	
	return threats
}

// hasFourInARow checks if player has a 4-in-a-row on the board
func (ai *AI) hasFourInARow(board *Board, player int) bool {
	// Define directions for line checking: horizontal, vertical, diagonal
	directions := []struct{ dr, dc int }{
		{0, 1},   // horizontal
		{1, 0},   // vertical
		{1, 1},   // diagonal \
		{1, -1},  // diagonal /
	}
	
	// Check entire board
	for row := 0; row < Size; row++ {
		for col := 0; col < Size; col++ {
			if board.grid[row][col] != player {
				continue
			}
			
			// Check all directions from this cell
			for _, dir := range directions {
				count := 1 // Start with 1 for the current cell
				
				// Look ahead in this direction
				for i := 1; i < 4; i++ {
					r, c := row+dir.dr*i, col+dir.dc*i
					if r < 0 || r >= Size || c < 0 || c >= Size || board.grid[r][c] != player {
						break
					}
					count++
				}
				
				if count == 4 {
					// Check if this sequence is open (can be extended to 5)
					r1, c1 := row-dir.dr, col-dir.dc
					r2, c2 := row+dir.dr*4, col+dir.dc*4
					
					// If either end is empty and on the board, it's an open 4 (very dangerous)
					isOpen := (r1 >= 0 && r1 < Size && c1 >= 0 && c1 < Size && board.grid[r1][c1] == Empty) ||
							  (r2 >= 0 && r2 < Size && c2 >= 0 && c2 < Size && board.grid[r2][c2] == Empty)
					
					if isOpen {
						return true
					}
				}
			}
		}
	}
	
	return false
}

// evaluateAllPatterns combines board evaluation and pattern recognition in a single pass
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
					dirScore += score
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

		// Enhanced threat detection for patterns
		// Add more weight to defensive moves that block a potential three-in-a-row
		if threats[move] >= threeInARow {
			// Check if this is a critical defensive position
			tempBoard := board.Copy()
			tempBoard.grid[move.Row][move.Col] = Player
			
			// Additional check for patterns that might lead to a win
			linePatterns := ai.getLinePatterns(tempBoard, move.Row, move.Col, Player)
			for pattern := range linePatterns {
				// Add additional score for defensive moves that block potential threats
				if pattern == "OOO_" || pattern == "_OOO" || pattern == "OO_O" || pattern == "O_OO" {
					threats[move] += fourInARow / 2 // Make it more important but not as critical as a direct 4-in-a-row
				}
			}
		}

		offensiveScores[move] = offensiveScore
		tempBoard.grid[move.Row][move.Col] = Empty
	}

	return threats, favors, offensiveScores
}

// extracts patterns from all lines passing through a point
func (ai *AI) getLinePatterns(board *Board, row, col, player int) map[string]bool {
	patterns := make(map[string]bool)
	
	// Define the 8 directions to check
	directions := []struct{ dr, dc int }{
		{-1, 0}, {1, 0},   // vertical
		{0, -1}, {0, 1},   // horizontal
		{-1, -1}, {1, 1},  // diagonal \
		{-1, 1}, {1, -1},  // diagonal /
	}
	
	// Check each of the 4 lines passing through this point
	for d := 0; d < 4; d++ {
		line := ""
		
		// Add characters for 4 spaces in one direction
		for i := 1; i <= 4; i++ {
			r := row + directions[d*2].dr * i
			c := col + directions[d*2].dc * i
			if r >= 0 && r < Size && c >= 0 && c < Size {
				if board.grid[r][c] == player {
					line += "O"  // Use O regardless of actual player (pattern matching convention)
				} else if board.grid[r][c] == Empty {
					line += "_"
				} else {
					line += "X"  // X is the opponent
				}
			}
		}
		
		// Add the central point
		centerChar := "O" // Player's piece
		
		// Add characters for 4 spaces in opposite direction
		oppLine := ""
		for i := 1; i <= 4; i++ {
			r := row + directions[d*2+1].dr * i
			c := col + directions[d*2+1].dc * i
			if r >= 0 && r < Size && c >= 0 && c < Size {
				if board.grid[r][c] == player {
					oppLine = "O" + oppLine  // Prepend
				} else if board.grid[r][c] == Empty {
					oppLine = "_" + oppLine  // Prepend
				} else {
					oppLine = "X" + oppLine  // Prepend
				}
			}
		}
		
		// Combine the line
		fullLine := oppLine + centerChar + line
		
		// Extract all 5-character patterns from this line
		for i := 0; i <= len(fullLine) - 5; i++ {
			pattern := fullLine[i:i+5]
			patterns[pattern] = true
		}
	}
	
	return patterns
}

// containsPattern checks if a string contains a specific pattern
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


// scans a line on the board for patterns
func (ai *AI) processLine(board *Board, startCol, startRow, dCol, dRow int, aiPatterns, playerPatterns map[string]int) {
	var line string
	var current int

	for i := 0; i < Size; i++ {
		row, col := startRow+i*dRow, startCol+i*dCol
		if row < 0 || row >= Size || col < 0 || col >= Size {
			break
		}

		current = board.grid[row][col]
		if current == AIPlayer {
			line += "X"
		} else if current == Player {
			line += "O"
		} else {
			line += "_"
		}
	}

	// Check for AI patterns
	for pattern := range ai.patternScores {
		aiLine := line
		count := countPattern(aiLine, pattern)
		if count > 0 {
			aiPatterns[pattern] += count
		}

		// Convert pattern for opponent (X->O)
		playerPattern := convertPattern(pattern)
		count = countPattern(line, playerPattern)
		if count > 0 {
			playerPatterns[pattern] += count // Store with original pattern name for weight lookup
		}
	}
}

// counts occurrences of a pattern in a string
func countPattern(line, pattern string) int {
	count := 0
	for i := 0; i <= len(line)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if pattern[j] == 'X' && line[i+j] != 'X' {
				match = false
				break
			} else if pattern[j] == 'O' && line[i+j] != 'O' {
				match = false
				break
			} else if pattern[j] == '_' && line[i+j] != '_' {
				match = false
				break
			}
		}
		if match {
			count++
		}
	}
	return count
}

// converts a pattern from AI perspective to player perspective
func convertPattern(pattern string) string {
	result := ""
	for _, ch := range pattern {
		if ch == 'X' {
			result += "O"
		} else if ch == 'O' {
			result += "X"
		} else {
			result += string(ch)
		}
	}
	return result
}

// adds bonuses/penalties based on stone positions
func evaluatePositionalFactors(board *Board) int {
	score := 0
	centerX, centerY := Size/2, Size/2

	// Center control bonus
	for r := centerY - 3; r <= centerY+3; r++ {
		for c := centerX - 3; c <= centerX+3; c++ {
			if r >= 0 && r < Size && c >= 0 && c < Size {
				// Distance from center (0 at center, increases outward)
				distance := max(abs(r-centerY), abs(c-centerX))

				if board.grid[r][c] == AIPlayer {
					// Bonus decreases with distance from center
					score += 30 - 5*distance
				} else if board.grid[r][c] == Player {
					// Penalty for opponent controlling center
					score -= 20 - 3*distance
				}
			}
		}
	}

	// Stone connectivity bonus - reward connected stones
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			if board.grid[r][c] == AIPlayer {
				// Count adjacent friendly stones (8 directions)
				adjacentStones := 0
				for dr := -1; dr <= 1; dr++ {
					for dc := -1; dc <= 1; dc++ {
						if dr == 0 && dc == 0 {
							continue // Skip the center cell
						}
						nr, nc := r+dr, c+dc
						if nr >= 0 && nr < Size && nc >= 0 && nc < Size && board.grid[nr][nc] == AIPlayer {
							adjacentStones++
						}
					}
				}
				// Bonus for connected stones (quadratic to reward clusters)
				score += adjacentStones * adjacentStones * 5
			}
		}
	}

	// Edge penalty - slight penalty for stones at board edges
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			if board.grid[r][c] != Empty {
				// Check if stone is at the edge
				isEdge := r == 0 || r == Size-1 || c == 0 || c == Size-1
				if isEdge {
					if board.grid[r][c] == AIPlayer {
						score -= 15 // Penalty for own stones at edge
					} else {
						score += 10 // Slight bonus for opponent's stones at edge
					}
				}
			}
		}
	}

	return score
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
