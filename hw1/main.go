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

const PRINT_MOVES = true // Print the moves made by the agent
const WALL_VALUE = 9001

type InitialState struct {
	X0            int
	Y0            int
	Battery       int
	MovementCost  int
	VacuumingCost int
	Tiles         [][]string
}

// Parse the initial state from a CSV file
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
		// First five lines are settings
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

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: clean.exe <algorithm('greedy'|'optimal')> <input csv file>")
		return
	}

	algorithm, filePath := os.Args[1], os.Args[2]
	initialState, err := ReadInitialState(filePath)
	if err != nil {
		log.Fatal(err)
	}

	if algorithm == "greedy" {
		FindAndTraverseGreedyPath(initialState)

	} else if algorithm == "optimal" {
		FindAndTraverseOptimalPath(initialState)

	} else {
		fmt.Println("Invalid algorithm")
		return
	}
}
