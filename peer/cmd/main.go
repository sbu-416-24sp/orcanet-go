package main

import (
	"flag"
	orcaBlockchain "orca-peer/internal/blockchain"
	orcaCLI "orca-peer/internal/cli"
	orcaHash "orca-peer/internal/hash"
	"os"
)

var boostrapNodeAddress string

func main() {
	flag.StringVar(&boostrapNodeAddress, "bootstrap", "", "Give address to boostrap.")
	flag.Parse()
	publicKey, privateKey := orcaHash.LoadInKeys()
	os.MkdirAll("./files/stored/", 0755)
	orcaBlockchain.StartBitcoinNode()
	orcaCLI.StartCLI(&boostrapNodeAddress, publicKey, privateKey)
}
