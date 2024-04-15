package main

import (
	"flag"
	orcaCLI "orca-peer/internal/cli"
	orcaHash "orca-peer/internal/hash"
	orcaTest "orca-peer/test"
	"os"
)

var test bool
var localOnly bool
var boostrapNodeAddress string

func main() {
	flag.BoolVar(&test, "test", false, "Create test server with no CLI.")
	flag.StringVar(&boostrapNodeAddress, "bootstrap", "", "Give address to boostrap.")
	flag.Parse()
	publicKey, privateKey := orcaHash.LoadInKeys()
	os.MkdirAll("./files/stored/", 0755)
	if test {
		orcaTest.RunTestServer()
	} else {
		orcaCLI.StartCLI(&boostrapNodeAddress, publicKey, privateKey)
	}
}
