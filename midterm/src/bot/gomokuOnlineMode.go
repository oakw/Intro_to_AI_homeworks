package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type GomokuOnlineClient struct {
	http      *http.Client
	aiX       int
	aiY       int
	onlineX   int
	onlineY   int
	gameState int
}


func initGomokuOnlineClient() *GomokuOnlineClient {
	gomokuOnlineClient := GomokuOnlineClient{}
	onlineStarts := rand.Intn(2) == 0
	cookieJar, _ := cookiejar.New(nil)
	gomokuOnlineClient.http = &http.Client{
		Jar: cookieJar,
	}

	r, err := gomokuOnlineClient.http.Get("https://gomokuonline.com/gomoku?reset=true&play=" + fmt.Sprintf("%t", onlineStarts) + "&random=" + fmt.Sprint(time.Now().Unix()))
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	cookieJar.SetCookies(r.Request.URL, r.Cookies())

	if err != nil {
		panic(err)

	} else if !onlineStarts && !bytes.Equal(bytes.Trim(body, "\""), []byte("ok")) {
		panic("Failed to start the game, returned: " + string(body))

	} else if !onlineStarts {
		// all ok, AI starts
		return &gomokuOnlineClient
	}

	var response []int
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Failed to parse response: ", err, string(body))
		panic(err)
	}

	gomokuOnlineClient.onlineX = response[1]
	gomokuOnlineClient.onlineY = response[2]
	gomokuOnlineClient.gameState = -1

	return &gomokuOnlineClient
}

func (client *GomokuOnlineClient) makeMoveAndObserve() GomokuOnlineClient {
	url := fmt.Sprintf("https://gomokuonline.com/gomoku?x=%d&y=%d&random=%d", client.aiX, client.aiY, time.Now().Unix())
	r, err := client.http.Get(url)
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	fmt.Println("Response: ", string(body))

	if err != nil {
		fmt.Println("Failed to read response: ", err)
		panic(err)
	}

	var response []int
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Failed to parse response: ", err)
		panic(err)
	}

	client.onlineX = response[1]
	client.onlineY = response[2]
	client.gameState = response[0]

	return *client
}


func RunGomokuOnlineMode(gameManager *GameManager) {
	client := initGomokuOnlineClient()
	if client == nil {
		fmt.Println("Failed to initialize online client")
		return
	}

	// Stream initial state if streaming is enabled
	if gameManager != nil && gameManager.streamEnabled {
		gameManager.StreamState()
	}

	for {
		if client.gameState == -1 {
			client.gameState = 0
		} else {
			client.makeMoveAndObserve()
		}

		if gameManager.board.MakeMove(client.onlineX, client.onlineY, Player) {
			// Update streaming info
			if gameManager != nil && gameManager.streamEnabled {
				gameManager.lastMoveRow = client.onlineX
				gameManager.lastMoveCol = client.onlineY
				gameManager.lastMoveBy = Player
				gameManager.turn = AIPlayer
				gameManager.StreamState()
			}

			if client.gameState > 0 {
				if gameManager != nil && gameManager.streamEnabled {
					gameManager.board.gameOver = true
					if gameManager.board.CheckWin(Player) {
						gameManager.winner = Player
					} else {
						gameManager.winner = AIPlayer
					}
					gameManager.StreamState()
				}
				break
			}

			// AI move
			bestMove := gameManager.ai.NextMove(gameManager.board, 3)
			gameManager.board.MakeMove(bestMove.Row, bestMove.Col, AIPlayer)

			// Update streaming info
			if gameManager.streamEnabled {
				gameManager.lastMoveRow = bestMove.Row
				gameManager.lastMoveCol = bestMove.Col
				gameManager.lastMoveBy = AIPlayer
				gameManager.turn = Player
				gameManager.StreamState()
			}

			client.aiX = bestMove.Row
			client.aiY = bestMove.Col
		}

		time.Sleep(100 * time.Millisecond)
	}
}
