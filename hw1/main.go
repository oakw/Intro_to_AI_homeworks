package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

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

	return Agent{initialState.X0, initialState.Y0, tiles, initialState.Battery, initialState.MovementCost, initialState.VacuumingCost}, nil
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
}

func (agent Agent) getTileValue(x int, y int) int {
	if x >= 0 && y >= 0 && y < len(agent.tiles) && y < len(agent.tiles[y]) {
		return agent.tiles[y][x]

	} else {
		return WALL_VALUE
	}
}

func (agent *Agent) currentTile() int {
	return agent.getTileValue(agent.posX, agent.posY)
}

func (agent *Agent) moveLeft() {
	if agent.posX > 0 {
		agent.posX--
	}
}

func (agent *Agent) moveRight() {
	if agent.posX < len(agent.tiles[agent.posY])-1 {
		agent.posX++
	}
}

func (agent *Agent) moveUp() {
	if agent.posY > 0 {
		agent.posY--
	}
}

func (agent *Agent) moveDown() {
	if agent.posY < len(agent.tiles)-1 {
		agent.posY++
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

	agent, err := CreateAgent(initialState)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(agent.currentTile())
	agent.moveRight()
	fmt.Println(agent.currentTile())
	agent.moveDown()
	fmt.Println(agent.currentTile())
	agent.moveDown()
	fmt.Println(agent.currentTile())
	agent.moveDown()
	fmt.Println(agent.currentTile())
	agent.moveRight()
	fmt.Println(agent.currentTile())

}
