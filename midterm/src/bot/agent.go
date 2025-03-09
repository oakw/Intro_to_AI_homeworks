package main

import (
	"fmt"
	"math"
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
}

func NewAI() *AI {
	return &AI{
		moveCount:               0,
		transpositionTable:      make(map[string]TTEntry),
		maxExecutionTimeSeconds: DEFAULT_MAX_MINIMAX_EXECUTION_SECONDS,
	}
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

	if ai.moveCount == 1 || len(board.moves) == 0 {
		// First move, play anywhere
		return possibleMoves[0]
	}

	threats, favors := board.evaluatePossibleThreatsAndFavors()

	possibleThreatMoves := make([]Move, len(possibleMoves))
	possibleFavorableMoves := make([]Move, len(possibleMoves))
	copy(possibleThreatMoves, possibleMoves)
	copy(possibleFavorableMoves, possibleMoves)
	fmt.Println("")

	sort.Slice(possibleThreatMoves, func(i, j int) bool {
		return threats[possibleThreatMoves[i]] > threats[possibleThreatMoves[j]]
	})

	sort.Slice(possibleFavorableMoves, func(i, j int) bool {
		return favors[possibleFavorableMoves[i]] > favors[possibleFavorableMoves[j]]
	})

	type PossibleMove struct {
		move  Move
		score int
	}

	movesFromEachType := []PossibleMove{}

	fmt.Println("Top threat move", possibleThreatMoves[0], "score", threats[possibleThreatMoves[0]])
	fmt.Println("Top favorable move", possibleFavorableMoves[0], "score", favors[possibleFavorableMoves[0]])

	if len(possibleFavorableMoves) > 0 {
		movesFromEachType = append(movesFromEachType, PossibleMove{move: possibleFavorableMoves[0], score: favors[possibleFavorableMoves[0]] + 1})

		if favors[possibleFavorableMoves[0]] >= fiveInARow {
			fmt.Println("Did favorable move", possibleFavorableMoves[0], "score", favors[possibleFavorableMoves[0]])
			return possibleFavorableMoves[0]
		}
	}

	if len(possibleThreatMoves) > 0 {
		movesFromEachType = append(movesFromEachType, PossibleMove{move: possibleThreatMoves[0], score: threats[possibleThreatMoves[0]]})

		if threats[possibleThreatMoves[0]] >= fiveInARow {
			fmt.Println("Reverse threat move", possibleThreatMoves[0], "score", threats[possibleThreatMoves[0]])
			return possibleThreatMoves[0]
		}
	}

	newBoard := board.Copy()
	ai.lastMinimaxExecutionStartTime = time.Now()
	ai.nodesExplored = 0
	ai.transpositionTable = make(map[string]TTEntry) // Clear TT for new search

	maxDepth := depth
	var bestMove Move = possibleMoves[0] // default in case we run out of time

	for currentDepth := 2; currentDepth <= maxDepth; currentDepth++ {
		rootNode := MinimaxNode{board: *newBoard, lastMove: board.moves[len(board.moves)-1], nextMove: Move{-1, -1}, currentTurn: AIPlayer}
		score := ai.minimax(&rootNode, currentDepth, alpha, beta)

		// Stop if we're close to time limit
		if time.Since(ai.lastMinimaxExecutionStartTime).Seconds() > ai.maxExecutionTimeSeconds*0.9 {
			break
		}

		bestMove = rootNode.nextMove
		fmt.Printf("Depth %d complete, best move: %v, score: %.2f, nodes: %d\n",
			currentDepth, bestMove, score, ai.nodesExplored)
	}

	movesFromEachType = append(movesFromEachType, PossibleMove{move: bestMove, score: int(float64(1))})
	fmt.Println("Minimax move", bestMove)

	sort.Slice(movesFromEachType, func(i, j int) bool {
		return movesFromEachType[i].score > movesFromEachType[j].score
	})

	if movesFromEachType[0].score < threeInARow && len(possibleFavorableMoves) > 0 {
		return possibleFavorableMoves[0]
	}

	return movesFromEachType[0].move
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
		return float64(node.board.evaluate())
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
