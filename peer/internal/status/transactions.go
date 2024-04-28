package status

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Data struct {
	Bytes               []byte `json:"bytes"`
	UnlockedTransaction []byte `json:"transaction"`
	PublicKey           string `json:"public_key"`
}

type Transaction struct {
	TransactionData Data   `json:"transaction"`
	TransactionDate string `json:"date"`
}

var AllTransactions []Transaction

func LoadInTransactions() {
	folderPath := "./files/transactions"
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return
	}
	if _, err := os.Stat("./files/transactions/transactions.json"); !os.IsNotExist(err) {
		// Open the file
		file, err := os.Open("./files/transactions/transactions.json")
		if err != nil {
			return
		}
		defer file.Close()
		fileData, err := io.ReadAll(file)
		if err != nil {
			return
		}
		var tempTransactions []Transaction
		err = json.Unmarshal(fileData, &tempTransactions)
		if err != nil {
			return
		}
		AllTransactions = tempTransactions
		return
	}

	var AllTransactions []Transaction
	fileInfos, err := os.ReadDir(folderPath)
	if err != nil {
		return
	}
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			filePath := filepath.Join(folderPath, fileInfo.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}
			var fileData Data
			if err := json.Unmarshal(content, &fileData); err != nil {
				fmt.Println("Error unmarshaling data:", err)
				continue
			}
			var transaction Transaction
			transaction.TransactionData = fileData
			transaction.TransactionDate = fileInfo.Name()
			AllTransactions = append(AllTransactions, transaction)
		}
	}
}

func CompressTransactions() error {
	folderPath := "./files/transactions"
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return err
	}

	var DirectoryTransactions []Transaction
	fileInfos, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && fileInfo.Name() != "transactions.json" {
			filePath := filepath.Join(folderPath, fileInfo.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}
			var fileData Data
			if err := json.Unmarshal(content, &fileData); err != nil {
				fmt.Println("Error unmarshaling data:", err)
				continue
			}
			var transaction Transaction
			transaction.TransactionData = fileData
			transaction.TransactionDate = fileInfo.Name()
			DirectoryTransactions = append(DirectoryTransactions, transaction)
		}
	}
	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error accessing path:", err)
			return nil
		}
		if !info.IsDir() {
			err := os.Remove(path)
			if err != nil {
				fmt.Println("Error removing file:", err)
				return nil
			}
			fmt.Println("Removed file:", path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(DirectoryTransactions)
	if err != nil {
		return err
	}
	file, err := os.Create("./files/transactions/transactions.json")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}
	return nil
}
