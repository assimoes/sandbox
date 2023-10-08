package main

import (
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type     string                 `json:"type"`
	Source   string                 `json:"source"`
	Target   string                 `json:"target"`
	Metadata map[string]interface{} `json:"metadata"`
}

var wsURL = "ws://localhost:8899/ws"

func main() {
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		panic("Cannot connect: " + err.Error())
	}
	defer c.Close()

	for {
		msg := Message{
			Type:     "movement",
			Source:   "A",
			Target:   "C",
			Metadata: map[string]interface{}{"timestamp": time.Now().Unix()},
		}

		err := c.WriteJSON(msg)
		if err != nil {
			panic("Write error: " + err.Error())
		}

		time.Sleep(5 * time.Second)
	}
}
