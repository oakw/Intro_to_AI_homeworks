package main

import (
	"fmt"
	"log"
	"math/rand"
)

// A bit dummy traversal algorithm that moves the agent in a greedy way.
// It moves the agent to the closest most dirty cell and cleans it.
func FindAndTraverseGreedyPath(initialState InitialState) {
	agent, err := CreateAgent(initialState)
	if err != nil {
		log.Fatal(err)
	}

	agent.logs = append(agent.logs, fmt.Sprintf("Initial position: (%d, %d)", agent.posX, agent.posY))

	allActions := []func(){agent.moveLeft, agent.moveRight, agent.moveUp, agent.moveDown}
	bestAction := func() {}
	noBestMoveDirectionIndex := 0

	for agent.battery > 0 {
		allActionWeights := []int{agent.getLeftMoveValue(), agent.getRightMoveValue(), agent.getUpMoveValue(), agent.getDownMoveValue()}
		bestAction = nil
		bestActionWeight := 0

		for i, action := range allActions {
			if allActionWeights[i] == WALL_VALUE {
				continue
			}

			if allActionWeights[i] > bestActionWeight {
				bestAction = action
				bestActionWeight = allActionWeights[i]
			}
		}

		// if no preferred move, do any random allowed move
		if bestAction == nil {
			// give it 100 tries to find a move
			for i := 0; i < 100; i++ {
				j := i
				// to prevent going on loop, pick a random move 30% of the time
				if rand.Float32() < 0.3 {
					j += 1
				}

				if allActionWeights[(noBestMoveDirectionIndex+j)%4] != WALL_VALUE {
					noBestMoveDirectionIndex = (noBestMoveDirectionIndex + j) % 4
					bestAction = allActions[noBestMoveDirectionIndex]
					break
				}
			}
		}

		if bestAction == nil {
			fmt.Println("No moves left")
			return
		}

		bestAction()
		agent.vacuumIfDirty()

		if agent.allTilesCleaned() {
			agent.logs = append(agent.logs, fmt.Sprintf("All tiles cleaned"))
			break
		}
	}

	if !agent.allTilesCleaned() {
		agent.logs = append(agent.logs, fmt.Sprintf("Battery depleted"))
	}

	if PRINT_MOVES {
		for _, log := range agent.logs {
			fmt.Println(log)
		}
	}

	agent.printStatistics()
}

// BFS to find the nearest non-zero value (target).
// Uses breadth-first search as described in
// https://en.wikipedia.org/wiki/Breadth-first_search with some additions (e.g. distance tracking and path reconstruction).
// This could be optimized code-wise.
func findNearestValuable(agent *Agent) []func(*Agent) {
	queue := Queue{}
	visited := make(map[[2]int]bool)
	distance := make(map[[2]int]int)
	predecessor := make(map[[2]int][2]int)

	start := [2]int{agent.posY, agent.posX}
	queue.Enqueue(start)
	visited[start] = true
	distance[start] = 0

	var bestPos *[2]int
	bestValue := -1

	// BFS traversal
	for !queue.IsEmpty() {
		curr, _ := queue.Dequeue()
		y, x := curr[0], curr[1]
		currDist := distance[curr]

		if currDist > agent.battery {
			continue
		}

		// Check if this cell has a non-zero value (excluding the start cell)
		tileValue := agent.getTileValue(x, y)
		if (y != start[0] || x != start[1]) && tileValue > 0 {
			// Prioritize the highest value; if equal, prefer the closest
			if tileValue > bestValue || (tileValue == bestValue && (bestPos == nil || currDist < distance[*bestPos])) {
				bestValue = tileValue
				bestPos = &curr
			}
		}

		// explore neighbors
		directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
		for _, dir := range directions {
			ny, nx := y+dir[0], x+dir[1]
			next := [2]int{ny, nx}

			if agent.getTileValue(nx, ny) != WALL_VALUE && !visited[next] {
				visited[next] = true
				queue.Enqueue(next)
				predecessor[next] = curr
				distance[next] = currDist + agent.movementCost
			}
		}
	}

	// no valid target was found, return nil
	if bestPos == nil {
		return nil
	}

	path := []func(*Agent){}
	current := *bestPos
	for current != start {
		direction := [2]int{current[0] - predecessor[current][0], current[1] - predecessor[current][1]}
		path = append(path, directionArrayToAction(direction))
		current = predecessor[current]
	}

	// reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

// FindAndTraverseOptimalPath finds the optimal path to clean all tiles by assuming task goals:
// Primary Goal: Clean as much dirt as possible.
// Secondary Goal: Clear (visit and clean) as many squares as possible.
// It combines BFS to find the nearest valuable cell and greedy actions to clean the dirt around the agent.
func FindAndTraverseOptimalPath(initialState InitialState) {
	agent, err := CreateAgent(initialState)
	if err != nil {
		log.Fatal(err)
	}

	agent.vacuumIfDirty()

	for agent.battery > 0 {
		var bestNext *[2]int
		bestVal := -1
		var bestAction func(*Agent)

		// check adjacent cells acting greedy
		directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
		for _, dir := range directions {
			ny, nx := agent.posY+dir[0], agent.posX+dir[1]
			tileValue := agent.getTileValue(nx, ny)
			if tileValue != WALL_VALUE && tileValue > bestVal {
				bestVal = tileValue
				bestNext = &[2]int{ny, nx}
				bestAction = directionArrayToAction(dir)
			}
		}

		if bestNext != nil && bestVal > 0 {
			// move to the best adjacent cell
			bestAction(&agent)
			agent.vacuumIfDirty()

		} else {
			// if no good adjacent cell, use BFS to find the nearest valuable cell
			pathToNode := findNearestValuable(&agent)
			if pathToNode == nil {
				break // no reachable non-zero tile, end
			}

			for _, pos := range pathToNode {
				pos(&agent)
				agent.vacuumIfDirty()

				if agent.battery <= 0 {
					break
				}
			}
		}
	}

	if PRINT_MOVES {
		for _, log := range agent.logs {
			fmt.Println(log)
		}
	}

	agent.printStatistics()
}
