package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // connected clients
var lock = sync.RWMutex{}                    // lock for concurrent access to the clients map

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error during WebSocket upgrade:", err)
		return
	}

	// Register the new client
	lock.Lock()
	clients[ws] = true
	lock.Unlock()

	fmt.Println("Client connected")

	defer func() {
		lock.Lock()
		delete(clients, ws)
		lock.Unlock()
		ws.Close()
		fmt.Println("Client disconnected")
	}()

	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Error while reading message:", err)
			break
		}

		// Broadcast the message to all connected clients
		lock.RLock()
		for client := range clients {
			if err := client.WriteMessage(messageType, p); err != nil {
				fmt.Println("Error while writing message:", err)
				lock.Lock()
				delete(clients, client)
				lock.Unlock()
			}
		}
		lock.RUnlock()

		fmt.Println("Received and broadcasted message")
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	fmt.Println("WebSocket server started on :8081")
	err := http.ListenAndServe(":8899", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}
