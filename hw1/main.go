package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const PRINT_MOVES = true

// https://stackoverflow.com/a/58841827
func ReadInitialState(filePath string) (InitialState, error) {
	initialState := InitialState{}

	f, err := os.Open(filePath)
	if err != nil {
		return initialState, errors.New(fmt.Sprintf("Error opening file %s: %v", filePath, err))
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comment = '#'
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields per record

	for i := 0; i < 5; i++ {
		setting, err := csvReader.Read()
		if err != nil {
			return initialState, errors.New(fmt.Sprintf("Error reading file %s: %v", filePath, err))
		}

		value, err := strconv.Atoi(strings.TrimSpace(strings.Split(setting[0], "#")[0]))
		if err != nil {
			return initialState, errors.New(fmt.Sprintf("Error parsing settings from file %s: %v", filePath, err))
		}

		if i == 0 {
			initialState.X0 = value
		}
		if i == 1 {
			initialState.Y0 = value
		}
		if i == 2 {
			initialState.Battery = value
		}
		if i == 3 {
			initialState.MovementCost = value
		}
		if i == 4 {
			initialState.VacuumingCost = value
		}
	}

	initialState.Tiles, err = csvReader.ReadAll()
	if err != nil {
		return initialState, errors.New(fmt.Sprintf("Error reading file %s: %v", filePath, err))
	}

	return initialState, nil
}

func CreateAgent(initialState InitialState) (Agent, error) {
	// Parse tiles from strings to integers
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

const WALL_VALUE = 9001

type InitialState struct {
	X0            int
	Y0            int
	Battery       int
	MovementCost  int
	VacuumingCost int
	Tiles         [][]string
}
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

func (agent Agent) getTileValue(x int, y int) int {
	if x >= 0 && y >= 0 && y < len(agent.tiles) && x < len(agent.tiles[y]) {
		return agent.tiles[y][x]

	} else {
		return WALL_VALUE
	}
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

func (agent *Agent) vacuumIfDirty() {
	dirtOnTile := agent.currentTile()
	if dirtOnTile > 0 && dirtOnTile < 9001 && agent.battery >= agent.vacuumingCost {
		agent.battery -= agent.vacuumingCost
		agent.tiles[agent.posY][agent.posX] = 0
		agent.dirtCleaned += dirtOnTile

		agent.logs = append(agent.logs,
			fmt.Sprintf("Vacuumed tile at (%d, %d), cleaned (%d) dirt", agent.posX, agent.posY, dirtOnTile))
	}
}

func (agent *Agent) moveBy(x int, y int) {
	if agent.battery >= agent.movementCost {
		agent.posX += x
		agent.posY += y
		agent.battery -= agent.movementCost

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

func FindGreedyPath(initialState InitialState) {
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

		// If no preferred move, do any random allowed move
		if bestAction == nil {
            // Give it 100 tries to find a move
			for i := 0; i < 100; i++ {
				j := i
				// To prevent going on loop, pick a random move 30% of the time
				if rand.Float32() < 0.3 {
					j+= 1
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
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <input csv file>")
		return
	}

	filePath := os.Args[1]
	initialState, err := ReadInitialState(filePath)
	if err != nil {
		log.Fatal(err)
	}

	FindGreedyPath(initialState)

	// agent, err := CreateAgent(initialState)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(agent.currentTile())
	// agent.moveRight()
	// fmt.Println(agent.currentTile())
	// agent.moveDown()
	// fmt.Println(agent.currentTile())
	// agent.moveDown()
	// fmt.Println(agent.currentTile())
	// agent.moveDown()
	// fmt.Println(agent.currentTile())
	// agent.moveRight()
	// fmt.Println(agent.currentTile())

}
