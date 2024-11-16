package process

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"example.com/model"
	"github.com/gorilla/websocket"
)

func ProcessChunk(conn *websocket.Conn, message []byte) error {
	//TODO: Read first 4 bytes
	metadataSize := binary.BigEndian.Uint32(message[:4])

	//TODO: Extract metadata
	metadataJSON := message[4 : 4+metadataSize]
	var metadata model.Metadata
	err := json.Unmarshal(metadataJSON, &metadata)
	if err != nil {
		fmt.Println("Error unmarshal metadata from a chunk")
		return err
	}

	//TODO: Extract binary data
	binaryData := message[4+metadataSize:]

	//TODO: Write chunk data to a tmp file
	err = WriteToTempFile(metadata, binaryData)
	if err != nil {
		return err
	}
	//TODO: Merge all tmp files to a complete WAV file
	if metadata.ChunkIndex == metadata.TotalChunks-1 {
		err = MergeToCompleteFile(metadata)
		if err != nil {
			return err
		}
	}

	return nil
}

func WriteToTempFile(metadata model.Metadata, binaryData []byte) error {
	folderName := "./websocket-temp"
	fileName := fmt.Sprintf("%s-%d.tmp", metadata.FileID, metadata.ChunkIndex)
	filePath := filepath.Join(folderName, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating temp file %s\n", filePath)
		return err
	}
	defer file.Close()

	_, err = file.Write(binaryData)
	if err != nil {
		fmt.Printf("Error writing binary data to file %s\n", filePath)
		return err
	}

	return nil
}

func MergeToCompleteFile(metadata model.Metadata) error {
	outputFolderName := "./websocket-wav"
	outputFileName := fmt.Sprintf("%s-server.wav", metadata.FileID)
	outputFilePath := filepath.Join(outputFolderName, outputFileName)

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("Error createing output wav file %s\n", outputFilePath)
		return err
	}
	defer outputFile.Close()

	for i := 0; i < metadata.TotalChunks; i++ {
		filePath := filepath.Join("./websocket-temp", fmt.Sprintf("%s-%d.tmp", metadata.FileID, i))
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s\n", filePath)
		}

		_, err = io.Copy(outputFile, file)
		if err != nil {
			fmt.Printf("Error writing chunk file %s to wav file %s\n", filePath, outputFilePath)
			return err
		}
		defer file.Close()

		os.Remove(filePath)
	}

	return nil
}
