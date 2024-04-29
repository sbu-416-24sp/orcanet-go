package main

import (
    "fmt"
    "errors"
    "net/http"
    "github.com/coloshword/OrcaNetAPIServer/manageOrcaNet"
) 

// startOrcaNet: starts an OrcaNet full node instance for the server to communicate with
func startOrcaNet() (error) {
   return manageOrcaNet.Start()
}

// startOrcaWallet: starts OrcoWallet instance for the server to communicate with
func startOrcaWallet() (error) {
    return manageOrcaNet.StartOrcaWallet() 
}

func main() {
    http.HandleFunc("/", getRoot)
    http.HandleFunc("/hello", getHello)
    http.HandleFunc("/getBlockchainInfo", getBlockchainInfo)
    http.HandleFunc("/getNewAddress", getNewAddress)
    http.HandleFunc("/getBalance", getBalance)
    http.HandleFunc("/mine", mine)
    http.HandleFunc("/sendToAddress", sendToAddress)
    http.HandleFunc("/getPeerInfo", getPeerInfo)
    http.HandleFunc("/getBestBlock", getBestBlock)
    http.HandleFunc("/getBestBlockInfo", getBestBlockInfo)
    fmt.Println("starting orcanet")
    startOrcaNet()    
    startOrcaWallet()
    err := http.ListenAndServe(":3333", nil)
    if errors.Is(err, http.ErrServerClosed) {
        fmt.Println("server is closed")
    } else if err != nil {
        fmt.Printf("error starting the server %s\n ", err)
    }
}

