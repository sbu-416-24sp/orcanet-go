package main

import (
    "fmt"
    "io"
    "net/http"
    "github.com/coloshword/OrcaNetAPIServer/manageOrcaNet"
    "strings"
	"strconv"
	"encoding/json"
)

type Block struct {
    Hash   string `json:"hash"`
    Height int    `json:"height"`
}

// some basic endpoints

// getRoot: the root endpoint ('/')
func getRoot(w http.ResponseWriter, r *http.Request) {
    fmt.Println("got / request")
    io.WriteString(w, "This is the root of the API server")
}

// getHello: the hello endpoint
func getHello(w http.ResponseWriter, r *http.Request) {
    fmt.Println("got hello request")
    io.WriteString(w, "Hello, HTTP!\n")
}

// getBlockchainInfo: endpoint to get the blockchain info
func getBlockchainInfo(w http.ResponseWriter, r *http.Request) {
    fmt.Println("getBlockchainInfo request") 
    const command string = "getblockchaininfo"
    stdout, err :=  manageOrcaNet.CallBtcctlCmd(command)
    fmt.Println(err)
    // return the output of CallBtcctlCommand back to the querier 
    io.WriteString(w, stdout)
}
// getNewAddress: endpoint to get a new wallet address
// this wallet address can be used for mining rewards / sending / receiving transactions
// For security purposes, it is recommended to create a new address everytime 
func getNewAddress(w http.ResponseWriter, r *http.Request) {
    fmt.Println("getNewAddress request")
    const command string = "getnewaddress --wallet"
    stdout, err := manageOrcaNet.CallBtcctlCmd(command)
    fmt.Println(err)
    io.WriteString(w, stdout)
}

// getBalance: gets the balance of the wallet 
func getBalance(w http.ResponseWriter, r *http.Request) {
    fmt.Println("getBalance endpoint")
    const command string = "getbalance --wallet"
    stdout, err := manageOrcaNet.CallBtcctlCmd(command)
    fmt.Println(err)
    io.WriteString(w, stdout)
}
// getPeerInfo: gets the peer info 
func getPeerInfo(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Get peer endpoint")
    const command string = "getpeerinfo"
    stdout, err := manageOrcaNet.CallBtcctlCmd(command)
    fmt.Println(err)
    io.WriteString(w, stdout)
}

// getBestBlock: gets the best block
func getBestBlock(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Get best block endpoint")
    const command string = "getbestblock"
    stdout, err := manageOrcaNet.CallBtcctlCmd(command)
    fmt.Println(err)
    io.WriteString(w, stdout)
}

// getBestBlockInfo: gets the best block info
func getBestBlockInfo(w http.ResponseWriter, r *http.Request) {
    // get bestblock hash then run getblock on it
    fmt.Println("Get best block info endpoint")
    const command string = "getbestblock"
    stdout, err := manageOrcaNet.CallBtcctlCmd(command)
    fmt.Println(err)
    // get the hash of the best block
    var block Block

    // Unmarshal the JSON data into the block variable
    err = json.Unmarshal([]byte(stdout), &block)
    if err != nil {
        fmt.Println(err)
        http.Error(w, "Error accessing JSON data", http.StatusInternalServerError)
    }

    getBlockCommand := "getblock " + block.Hash
    stdout, err = manageOrcaNet.CallBtcctlCmd(getBlockCommand)
    fmt.Println(err)
    io.WriteString(w, stdout)
}
// mine: endpoint to start mining, mining rewards go to the associated wallet on this node 
func mine(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Mine endpoint")
    // to mine, we need to restart the OrcaNet node with --generate and --mingaddr=<newAddress>
    const getAddressCmd string = "getnewaddress --wallet"
    stdout, err := manageOrcaNet.CallBtcctlCmd(getAddressCmd)
    if err != nil {
        fmt.Println("error getting a new address for mining")
        io.WriteString(w, "error getting a new address for mining")
    }
    //std out is the address
    address := strings.TrimSpace(stdout)
    orcaNetParams := []string{"--generate", "--miningaddr=" + address}
    fmt.Println("orcaParams[0] " + orcaNetParams[0])
    fmt.Println("orcaParams[1] " + orcaNetParams[1])
    // we need to kill the OrcaNet process first
    if err := manageOrcaNet.Stop(); err != nil {
        fmt.Println("failed to end OrcaNet:", err)
        http.Error(w, "failed to kill original OrcaNet instance to start mining", http.StatusInternalServerError)
        return 
    }
    if err := manageOrcaNet.Start(orcaNetParams...); err != nil {
        fmt.Println("failed to start mining:", err)
        http.Error(w, "failed to start mining", http.StatusInternalServerError)
        return
    }
    io.WriteString(w, "Mining successfully started")
}

// sendToAddress: endpoint to send n coins to an address
// if you want to send coins to a specific wallet, ask the recepient to getNewAddress and pass that address to the query string 
// Usage: make a JSON request with 2 fields "coins" and "address"
func sendToAddress(w http.ResponseWriter, r *http.Request) {
    var request struct {
        Coins           string `json:"coins"`
        Address         string `json:"address"`
        SenderWalletPass string `json:"senderwalletpass"`
    }
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Error reading request body", http.StatusBadRequest)
        return
    }

    if request.Coins == "" || request.Address == "" || request.SenderWalletPass == "" {
        http.Error(w, "Missing fields in request", http.StatusBadRequest)
        return
    }

    if _, err := strconv.ParseFloat(request.Coins, 64); err != nil {
        http.Error(w, "Invalid number format for coins", http.StatusBadRequest)
        return
    }

    if err := unlockWallet(request.SenderWalletPass); err != nil {
        fmt.Printf("Failed to unlock wallet: %v\n", err)
        http.Error(w, "Error unlocking wallet", http.StatusInternalServerError)
        return
    }

    if err := sendCoins(request.Coins, request.Address); err != nil {
        fmt.Printf("Error sending coins: %v\n", err)
        http.Error(w, "Error sending coins", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Successfully sent %s coins to %s\n", request.Coins, request.Address)
}

func unlockWallet(walletPass string) error {
    command := fmt.Sprintf("--wallet walletpassphrase %s 100", walletPass)
    stdout, err := manageOrcaNet.CallBtcctlCmd(command)
    if err != nil {
        return fmt.Errorf("Failed to unlock wallet: %s, error: %v", stdout, err)
    }
    return nil
}

func sendCoins(numCoins, address string) error {
    command := fmt.Sprintf("--wallet sendtoaddress %s %s", address, numCoins)
    stdout, err := manageOrcaNet.CallBtcctlCmd(command)
    if err != nil {
        return fmt.Errorf("Failed to send coins: %s, error: %v", stdout, err)
    }
    return nil
}



// stopMine: endpoint to stop mining
func stopMine(w http.ResponseWriter, r *http.Request) {
    fmt.Println("stop mine endpoint")
}





