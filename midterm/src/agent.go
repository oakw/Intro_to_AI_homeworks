package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

type AI struct {
	moveCount int
}

func NewAI() *AI {
	return &AI{moveCount: 0}
}

func (ai *AI) NextMove(board *Board, depth int, alpha, beta float64, isMaximizing bool) Move {
	ai.moveCount += 1
	possibleMoves := GenerateMoves(board, AIPlayer)

	if ai.moveCount == 1 {
		// First move, play anywhere
		return possibleMoves[rand.Intn(len(possibleMoves))]
	}

	threats, _ := board.evaluatePossibleThreatsAndFavors()

	possibleThreatMoves := possibleMoves
	// possibleFavorableMoves := possibleMoves

	sort.Slice(possibleThreatMoves, func(i, j int) bool {
		return threats[possibleThreatMoves[i]] > threats[possibleThreatMoves[j]]
	})

	// sort.Slice(possibleFavorableMoves, func(i, j int) bool {
	// 	return favors[possibleFavorableMoves[i]] > favors[possibleFavorableMoves[j]]
	// })

	if len(possibleThreatMoves) > 0 && threats[possibleThreatMoves[0]] >= threeInARow {
		fmt.Println("Threat from opponent detected", possibleThreatMoves[0])
		return possibleThreatMoves[0]
	}

	// if len(possibleFavorableMoves) > 0 && favors[possibleFavorableMoves[0]] >= fourInARow {
	// 	fmt.Println("Favorable move detected", possibleFavorableMoves[0])
	// 	return possibleFavorableMoves[0]
	// }

	newBoard := board.Copy()
	rootNode := MinimaxNode{board: *newBoard, lastMove: board.moves[len(board.moves)-1], nextMove: Move{-1, -1}, currentTurn: AIPlayer}
	score := ai.minimax(&rootNode, depth, alpha, beta)
	fmt.Println("Score", score)

	return rootNode.nextMove
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
	if depth == 0 || node.board.CheckWin(AIPlayer) || node.board.CheckWin(Player) {
		return float64(node.board.evaluate())
	}

	if node.currentTurn == AIPlayer {
		maxEval := -math.MaxFloat64
		for _, childNode := range node.generateChildNodes() {
			eval := ai.minimax(&childNode, depth-1, alpha, beta)
			if eval > maxEval {
				maxEval = eval
				node.nextMove = childNode.lastMove
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
				node.nextMove = childNode.lastMove
			}
			beta = math.Min(beta, eval)
			if beta <= alpha {
				break
			}
		}
		return minEval
	}
}
