package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Agent struct {
	posX          int
	posY          int
	tiles         [][]int
	battery       int
	movementCost  int
	vacuumingCost int
	dirtCleaned   int
	tilesMoved    int
	logs          []string
}

func CreateAgent(initialState InitialState) (Agent, error) {
	// convert tiles from strings to integers
	tiles := [][]int{}
	for y, row := range initialState.Tiles {
		tiles = append(tiles, make([]int, len(row)))
		for x, tile := range row {
			val, err := strconv.Atoi(strings.TrimSpace(tile))
			if err != nil {
				return Agent{}, errors.New(fmt.Sprintf("Error parsing tile at (%d, %d): %v", x, y, err))
			}

			tiles[y][x] = val
		}
	}

	return Agent{
		posX:          initialState.X0,
		posY:          initialState.Y0,
		tiles:         tiles,
		battery:       initialState.Battery,
		movementCost:  initialState.MovementCost,
		vacuumingCost: initialState.VacuumingCost,
		dirtCleaned:   0,
		tilesMoved:    0,
		logs:          []string{},
	}, nil
}

func (agent Agent) getTileValue(x int, y int) int {
	if x >= 0 && y >= 0 && y < len(agent.tiles) && x < len(agent.tiles[y]) {
		return agent.tiles[y][x]

	} else {
		return WALL_VALUE
	}
}

func (agent *Agent) printStatistics() {
	fmt.Printf("Dirt cleaned: %d\n", agent.dirtCleaned)
	fmt.Printf("Tiles moved: %d\n", agent.tilesMoved)
	fmt.Printf("Battery remaining: %d\n", agent.battery)
}

func (agent *Agent) allTilesCleaned() bool {
	for _, row := range agent.tiles {
		for _, tile := range row {
			if tile > 0 && tile < 9001 {
				return false
			}
		}
	}

	return true
}

func (agent *Agent) currentTile() int {
	return agent.getTileValue(agent.posX, agent.posY)
}

func (agent *Agent) getLeftMoveValue() int {
	return agent.getTileValue(agent.posX-1, agent.posY)
}

func (agent *Agent) getRightMoveValue() int {
	return agent.getTileValue(agent.posX+1, agent.posY)
}

func (agent *Agent) getUpMoveValue() int {
	return agent.getTileValue(agent.posX, agent.posY-1)
}

func (agent *Agent) getDownMoveValue() int {
	return agent.getTileValue(agent.posX, agent.posY+1)
}

func (agent *Agent) moveBy(x int, y int) {
	if agent.battery >= agent.movementCost {
		agent.posX += x
		agent.posY += y
		agent.battery -= agent.movementCost
		agent.tilesMoved += 1

		agent.logs = append(agent.logs, fmt.Sprintf("Moved to (%d, %d)", agent.posX, agent.posY))
	}
}

func (agent *Agent) moveLeft() {
	if agent.getLeftMoveValue() != WALL_VALUE {
		agent.moveBy(-1, 0)
	}
}

func (agent *Agent) moveRight() {
	if agent.getRightMoveValue() != WALL_VALUE {
		agent.moveBy(1, 0)
	}
}

func (agent *Agent) moveUp() {
	if agent.getUpMoveValue() != WALL_VALUE {
		agent.moveBy(0, -1)
	}
}

func (agent *Agent) moveDown() {
	if agent.getDownMoveValue() != WALL_VALUE {
		agent.moveBy(0, 1)
	}
}

func (agent *Agent) vacuumIfDirty() int {
	dirtOnTile := agent.currentTile()
	if dirtOnTile > 0 && dirtOnTile < 9001 && agent.battery >= agent.vacuumingCost {
		agent.battery -= agent.vacuumingCost
		agent.tiles[agent.posY][agent.posX] = 0
		agent.dirtCleaned += dirtOnTile

		agent.logs = append(agent.logs,
			fmt.Sprintf("Vacuumed tile at (%d, %d), cleaned (%d) dirt", agent.posX, agent.posY, dirtOnTile))
		return dirtOnTile
	}

	return 0
}
