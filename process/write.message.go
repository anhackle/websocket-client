package process

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"example.com/model"
	"github.com/gorilla/websocket"
)

var (
	chunkSize = 1024
)

func SendFileToServer(conn *websocket.Conn, fileID, fileName string, chunkSize int) error {
	folderName := "./websocket-wav"
	filePath := filepath.Join(folderName, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s\n", filePath)
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	var totalChunks int
	if fileInfo.Size()%int64(chunkSize) != 0 {
		totalChunks = int(fileInfo.Size()/int64(chunkSize)) + 1
	} else {
		totalChunks = int(fileInfo.Size() / int64(chunkSize))
	}

	buffer := make([]byte, chunkSize)

	for chunkIndex := 0; ; chunkIndex++ {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Printf("Error reading content of file %s\n", filePath)
			return err
		}

		if n == 0 {
			break
		}

		metadata := model.Metadata{
			FileID:      fileID,
			ChunkIndex:  chunkIndex,
			TotalChunks: totalChunks,
		}
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			fmt.Printf("Error converting struct to json\n")
			return err
		}

		//TODO: Send message to websocket
		err = SendMessage(conn, metadataJSON, buffer, n)
		if err != nil {
			fmt.Printf("Error sending message throuh websocket\n")
			return err
		}
	}

	return nil
}

func SendMessage(conn *websocket.Conn, metadataJSON, buffer []byte, numberofBytes int) error {
	message := bytes.NewBuffer(nil)

	binary.Write(message, binary.BigEndian, int32(len(metadataJSON)))
	message.Write(metadataJSON)
	message.Write(buffer[:numberofBytes])

	err := conn.WriteMessage(websocket.BinaryMessage, message.Bytes())
	if err != nil {
		fmt.Printf("Error sending chunk through websocket\n")
		return err
	}

	return nil
}
