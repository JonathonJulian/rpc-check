package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type jsonRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  string `json:"result"`
}

type BlockHeightData struct {
	LocalHeight      int64
	HighestRefHeight int64
	Status           string
	Mutex            sync.RWMutex
}

var blockHeightData = BlockHeightData{}

func fetchBlockHeight(nodeURL string) (int64, error) {
	requestBody := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_blockNumber",
		Params:  []interface{}{},
		ID:      1,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(nodeURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response jsonRPCResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return 0, err
	}

	blockHeight, err := strconv.ParseInt(response.Result[2:], 16, 64)
	if err != nil {
		return 0, err
	}

	return blockHeight, nil
}

func updateBlockHeights(localNodeURL string, referenceNodeURLs []string) {
	for {
		localHeight, err := fetchBlockHeight(localNodeURL)
		if err != nil {
			log.Printf("Error fetching local node block height: %v. Marking as down.", err)
			blockHeightData.Mutex.Lock()
			blockHeightData.Status = "down\n"
			blockHeightData.Mutex.Unlock()
		} else {
			highestHeight := int64(0)
			for _, url := range referenceNodeURLs {
				height, err := fetchBlockHeight(url)
				if err != nil {
					log.Printf("Error fetching reference node block height: %v", err)
					continue
				}
				if height > highestHeight {
					highestHeight = height
				}
			}

			blockHeightData.Mutex.Lock()
			blockHeightData.LocalHeight = localHeight
			blockHeightData.HighestRefHeight = highestHeight
			if localHeight >= highestHeight {
				blockHeightData.Status = "up weight=100\n"
			} else {
				blockHeightData.Status = "up weight=50\n"
			}
			blockHeightData.Mutex.Unlock()

			if localHeight >= highestHeight {
				blockHeightData.Status = "up weight=100\n"
			} else {
				blockHeightData.Status = "up weight=50\n" // Adjust to 50% weight if out of sync
			}
		}
		time.Sleep(500 * time.Millisecond) // Update every 500 milliseconds
	}
}

func startAgentCheckServer() {
	listenPort := os.Getenv("AGENT_LISTEN_PORT")
	if listenPort == "" {
		listenPort = "9876" // Default port for HAProxy agent checks
	}

	listener, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		log.Fatalf("Failed to start agent check server: %v", err)
	}
	defer listener.Close()
	log.Printf("Agent check server listening on :%s\n", listenPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go func(c net.Conn) {
			defer func() {
				err := c.Close()
				if err != nil {
					log.Printf("Failed to close connection: %v", err)
				}
			}()

			blockHeightData.Mutex.RLock()
			status := blockHeightData.Status
			blockHeightData.Mutex.RUnlock()

			// Logging the status to be sent for transparency
			log.Printf("Sending status: %s", status)
			// For demonstration, hardcoded "weight=66\n" is being sent
			// Replace this with `status` if you want to send dynamic status based on block height
			_, err := c.Write([]byte("up 67%\n"))
			if err != nil {
				log.Printf("Failed to send response: %v", err)
				return
			}

			// Optionally log that the response was successfully sent
			log.Println("Response successfully sent, closing connection.")
		}(conn)
	}
}

func main() {
	localNodeURL := os.Getenv("LOCAL_NODE_URL")
	referenceNodeURLsEnv := os.Getenv("REFERENCE_NODE_URLS")
	if localNodeURL == "" || referenceNodeURLsEnv == "" {
		log.Fatal("LOCAL_NODE_URL and REFERENCE_NODE_URLS environment variables must be set.")
	}
	referenceNodeURLs := strings.Split(referenceNodeURLsEnv, ",")

	go updateBlockHeights(localNodeURL, referenceNodeURLs)
	startAgentCheckServer()
}
