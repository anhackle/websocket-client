package main

import (
	"crypto/tls"
	"fmt"

	"example.com/process"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	filePath  = "client.wav"
	chunkSize = 1024
	fileID    = uuid.New().String()
)

func handleConnection() (*websocket.Conn, error) {
	serverURL := "wss://localhost:8081/?callid=123456&agentid=123"

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, _, err := dialer.Dial(serverURL, nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket server:", err)
		return nil, err
	}

	return conn, nil

}

func main() {
	conn, err := handleConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	go func() {
		err := process.SendFileToServer(conn, fileID, filePath, chunkSize)
		if err != nil {
			fmt.Println("Error sending file:", err)
		}
	}()

	receiveMessageFromServer(conn)
}

func receiveMessageFromServer(conn *websocket.Conn) {
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		switch messageType {
		case websocket.TextMessage:
			if string(message) == "Close" {
				err := conn.Close()
				if err != nil {
					fmt.Println("Error close websocket connection", err)
					return
				}
				fmt.Println("Close websocket connection")
				return
			}
		case websocket.BinaryMessage:
			err := process.ProcessChunk(conn, message)
			if err != nil {
				fmt.Println("Error received file chunk from server", err)
				return
			}
		}
	}
}
