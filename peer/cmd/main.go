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
}
