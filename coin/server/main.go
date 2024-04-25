package main

import (
    "fmt"
    "errors"
    "os/signal"
    "os/exec"
    "os"
    "syscall"
    "net/http"
    "github.com/coloshword/OrcaNetAPIServer/manageOrcaNet"
) 

// startOrcaNet: starts an OrcaNet full node instance for the server to communicate with
func startOrcaNet() (*exec.Cmd, error) {
    proc, err := manageOrcaNet.Start()
    return proc, err
}

// startOrcaWallet: starts OrcoWallet instance for the server to communicate with
func startOrcaWallet() (*exec.Cmd, error) {
    proc, err := manageOrcaNet.StartOrcaWallet() 
    return proc, err
}

func main() {
    netProc, _ := startOrcaNet()    
    walletProc, _ := startOrcaWallet()
    go startHTTPServer()
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
    defer close(sig)

    <-sig
    if netProc != nil { 
        netProc.Process.Kill()
    }

    if walletProc != nil {
        walletProc.Process.Kill()
    }
}

func startHTTPServer(){
    http.HandleFunc("/", getRoot)
    http.HandleFunc("/hello", getHello)
    http.HandleFunc("/getBlockchainInfo", getBlockchainInfo)
    http.HandleFunc("/getNewAddress", getNewAddress)
    http.HandleFunc("/getBalance", getBalance)
    http.HandleFunc("/mine", mine)
    http.HandleFunc("/sendToAddress", sendToAddress)
    fmt.Println("starting orcanet")

    err := http.ListenAndServe(":3333", nil)
    if errors.Is(err, http.ErrServerClosed) {
        fmt.Println("server is closed")
    } else if err != nil {
        fmt.Printf("error starting the server %s\n ", err)
    }
}

