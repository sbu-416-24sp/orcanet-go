package main

import (
	"flag"
	"fmt"
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
	fmt.Println("**Starting Blockchain Server**")
	const executablePath string = "/Users/suryapatil/OrcaNetAPIServer/OrcaNetAPIServer"
	cmd := exec.Command(executablePath)
	cmd.Start()
	fmt.Println("**Starting Blockchain Server started**")
	orcaCLI.StartCLI(&boostrapNodeAddress, publicKey, privateKey)
}
