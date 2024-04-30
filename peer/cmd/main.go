package main

import (
	"flag"
	"fmt"
	orcaAPI "orca-peer/internal/api"
	orcaCLI "orca-peer/internal/cli"
	orcaHash "orca-peer/internal/hash"
	"os"
	"os/exec"
)

var boostrapNodeAddress string

func main() {
	flag.StringVar(&boostrapNodeAddress, "bootstrap", "", "Give address to boostrap.")
	flag.Parse()
	publicKey, privateKey := orcaHash.LoadInKeys()
	os.MkdirAll("./files/stored/", 0755)

	cmd := exec.Command("./OrcaNetAPIServer")
	cmd.Dir = "../coin/"
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting OrcaNetAPIServer: %s\n", err)
		return
	}
	fmt.Println("Started block chain api server")
	orcaCLI.StartCLI(&boostrapNodeAddress, publicKey, privateKey, cmd, orcaAPI.InitServer)
}
