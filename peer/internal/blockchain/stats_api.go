package blockchain

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	orcaHash "orca-peer/internal/hash"
	orcaStatus "orca-peer/internal/status"
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
func getLatestTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
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
