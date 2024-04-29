package blockchain

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	orcaHash "orca-peer/internal/hash"
	orcaStatus "orca-peer/internal/status"
	"os"
	"sort"
	"time"
)

var publicKey *rsa.PublicKey

type TransactionFile struct {
	Bytes               []byte  `json:"bytes"`
	UnlockedTransaction []byte  `json:"transaction"`
	PublicKey           string  `json:"public_key"`
	Date                string  `json:"date"`
	Cost                float64 `json:"cost"`
}

func getDailyRevenue(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pubKeyString, err := orcaHash.ExportRsaPublicKeyAsPemStr(publicKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		totalSent := 0
		totalReceived := 0
		orcaStatus.LoadInTransactions()
		timeThreshold := time.Now().Add(-24 * time.Hour)
		for _, transaction := range orcaStatus.AllTransactions {
			timestamp, err := time.Parse(time.RFC3339Nano, transaction.TransactionData.Date)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			if timestamp.After(timeThreshold) {
				if transaction.TransactionData.PublicKey == string(pubKeyString) {
					totalSent += int(transaction.TransactionData.Cost)
				} else {
					totalReceived += int(transaction.TransactionData.Cost)
				}
			}
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

type Revenue struct {
	Date     string `json:"date"`
	Earning  int    `json:"earning"`
	Spending int    `json:"spending"`
}

func getMonthlyRevenue(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pubKeyString, err := orcaHash.ExportRsaPublicKeyAsPemStr(publicKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		orcaStatus.LoadInTransactions()
		timeThreshold := time.Now().Add(-24 * 30 * time.Hour)
		hashMap := make(map[string]*Revenue)
		for _, transaction := range orcaStatus.AllTransactions {
			timestamp, err := time.Parse(time.RFC3339Nano, transaction.TransactionData.Date)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			if timestamp.After(timeThreshold) {
				key := timestamp.Month().String() + "/" + string(timestamp.Day())
				value, ok := hashMap[key]
				var rev *Revenue
				rev = nil
				if ok {
					rev = value
				} else {
					rev = &Revenue{Date: key, Earning: 0, Spending: 0}
					hashMap[key] = rev
				}
				if transaction.TransactionData.PublicKey == string(pubKeyString) {
					rev.Spending += int(transaction.TransactionData.Cost)
				} else {
					rev.Earning += int(transaction.TransactionData.Cost)
				}
			}
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
func getYearlyRevenue(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pubKeyString, err := orcaHash.ExportRsaPublicKeyAsPemStr(publicKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		orcaStatus.LoadInTransactions()
		timeThreshold := time.Now().Add(-24 * 365 * time.Hour)
		hashMap := make(map[string]*Revenue)
		for _, transaction := range orcaStatus.AllTransactions {
			timestamp, err := time.Parse(time.RFC3339Nano, transaction.TransactionData.Date)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			if timestamp.After(timeThreshold) {
				key := timestamp.Month().String() + "/" + string(timestamp.Day())
				value, ok := hashMap[key]
				var rev *Revenue
				rev = nil
				if ok {
					rev = value
				} else {
					rev = &Revenue{Date: key, Earning: 0, Spending: 0}
					hashMap[key] = rev
				}
				if transaction.TransactionData.PublicKey == string(pubKeyString) {
					rev.Spending += int(transaction.TransactionData.Cost)
				} else {
					rev.Earning += int(transaction.TransactionData.Cost)
				}
			}
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

type TransactionFileData struct {
	Bytes               []byte  `json:"bytes"`
	UnlockedTransaction []byte  `json:"transaction"`
	PublicKey           string  `json:"public_key"`
	Date                string  `json:"date"`
	Cost                float64 `json:"cost"`
}
type TransactionResponse struct {
	Id       string `json:"id"`
	Reciever string `json:"receiver"`
	Amount   string `json:"amount"`
	Status   string `json:"status"`
	Reason   string `json:"reason"`
	Date     string `json:"date"`
}
type Transaction struct {
	Price     float64 `json:"price"`
	Timestamp string  `json:"timestamp"`
	Uuid      string  `json:"uuid"`
}
type LatestTransactionResponse struct {
	WalletId     string                `json:"wallet_id"`
	Transactions []TransactionResponse `json:"transactions"`
}

func getLatestTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		dir := "./files/transactions"
		files, err := os.ReadDir(dir)
		if err != nil {
			http.Error(w, "error reading directory", http.StatusInternalServerError)
			return
		}
		type FileWithTime struct {
			Name     string
			Modified time.Time
		}
		var filesWithTime []FileWithTime
		for _, file := range files {
			t, err := time.Parse(time.RFC3339Nano, file.Name())
			if err != nil {
				continue
			}
			filesWithTime = append(filesWithTime, FileWithTime{Name: file.Name(), Modified: t})
		}
		sort.Slice(filesWithTime, func(i, j int) bool {
			return filesWithTime[i].Modified.Before(filesWithTime[j].Modified)
		})
		latestTransactions := make([]TransactionResponse, 0)
		for idx, file := range filesWithTime {
			if idx > 5 {
				break
			}
			file, err := os.Open("./files/transactions/" + file.Name)
			if err != nil {
				http.Error(w, "transaction file does not exist", http.StatusInternalServerError)
				return
			}
			defer file.Close()
			fileContent, err := io.ReadAll(file)
			if err != nil {
				http.Error(w, "Error reading in file", http.StatusInternalServerError)
				return
			}
			var data TransactionFileData
			err = json.Unmarshal(fileContent, &data)
			if err != nil {
				http.Error(w, "Error unmarshaling JSON", http.StatusInternalServerError)
				return
			}
			var transaction Transaction
			err = json.Unmarshal(data.UnlockedTransaction, &transaction)
			if err != nil {
				fmt.Println("Error unmarshalling JSON:", err)
				return
			}
			latestTransactions = append(latestTransactions, TransactionResponse{
				Id:       transaction.Uuid,
				Reciever: "",
				Amount:   fmt.Sprintf("%f", transaction.Price),
				Status:   "Success",
				Reason:   "Auto-Transaction",
				Date:     transaction.Timestamp,
			})
		}
		response := LatestTransactionResponse{
			WalletId:     "",
			Transactions: latestTransactions,
		}
		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonData)
		if err != nil {
			http.Error(w, "Error  marshalling JSON", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
func getCompleteTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
func getStatsNetwork(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
func InitBlockchainStats(publicKey *rsa.PublicKey) {
	http.HandleFunc("/wallet/revenue/daily", getDailyRevenue)
	http.HandleFunc("/wallet/revenue/monthly", getMonthlyRevenue)
	http.HandleFunc("/wallet/revenue/yearly", getYearlyRevenue)

	http.HandleFunc("/wallet/transactions/latest", getLatestTransactions)
	http.HandleFunc("/wallet/revenue/complete", getCompleteTransactions)

	http.HandleFunc("/wallet/revenue/yearly", getStatsNetwork)
}
