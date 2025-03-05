package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

const MAX_MINIMAX_EXECUTION_SECONDS = 2

type AI struct {
	moveCount                     int
	lastMinimaxExecutionStartTime time.Time
}

func NewAI() *AI {
	return &AI{moveCount: 0}
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
	rootNode := MinimaxNode{board: *newBoard, lastMove: board.moves[len(board.moves)-1], nextMove: Move{-1, -1}, currentTurn: AIPlayer}
	score := ai.minimax(&rootNode, depth, alpha, beta)
	movesFromEachType = append(movesFromEachType, PossibleMove{move: rootNode.nextMove, score: int(score) + 1})
	fmt.Println("Minimax move", rootNode.nextMove, "score", score)

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

func (node *MinimaxNode) generateChildNodes() []MinimaxNode {
	moves := GenerateMoves(&node.board, node.currentTurn)
	childNodes := make([]MinimaxNode, len(moves))
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
	if depth == 0 || node.board.CheckWin(AIPlayer) || node.board.CheckWin(Player) || time.Since(ai.lastMinimaxExecutionStartTime).Seconds() > MAX_MINIMAX_EXECUTION_SECONDS {
		return float64(node.board.evaluate())
	}

	if node.currentTurn == AIPlayer {
		maxEval := -math.MaxFloat64
		for _, childNode := range node.generateChildNodes() {
			eval := ai.minimax(&childNode, depth-1, alpha, beta)
			if eval > maxEval {
				maxEval = eval
				node.nextMove = Move{Row: childNode.lastMove.Row, Col: childNode.lastMove.Col}
			}
			alpha = math.Max(alpha, eval)
			if beta <= alpha {
				break
			}
		}
		return maxEval
	} else {
		minEval := math.MaxFloat64
		for _, childNode := range node.generateChildNodes() {
			eval := ai.minimax(&childNode, depth-1, alpha, beta)
			if eval < minEval {
				minEval = eval
				node.nextMove = Move{Row: childNode.lastMove.Row, Col: childNode.lastMove.Col}
			}
			beta = math.Min(beta, eval)
			if beta <= alpha {
				break
			}
		}
		return minEval
	}
}
