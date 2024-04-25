# API Server wrapper for OrcaNet(full node) and the OrcaWallet, and Btcctl 
## Overview 
Running this program creates an HTTP server, and also runs btcd and btcwallet in the background for you. Since it is an HTTP server, all you have to do is run the executable from your code, such as using `os/exec` in Go, or `subprocess` in Python. After that, just make requests from your language of choice.
### Usage 
IMPORTANT: If you previously ran btcd or another team's blockchain, you need to remove btcwallet and btcd folders, or you won't be able to connect to the bootstrap node. On MacOS they are in `Library/Application\ Support/`

1) run `./build.sh`. This builds the project and also prompts you to create a wallet if you don't have one. If you already have a wallet, it'll just tell you it already exists

2) Run the server executable `OrcaNetAPIServer`. You can do this in the terminal by running `./OrcaNetAPIServer`, but you probably want to run it as part of your program. Here's a simple example using Go:

```Go
import (
    "os/exec"
)

const executablePath string = <PATH TO OrcaNetAPIServer>
cmd := exec.Command(executablePath)
cmd.Start()

```

3) Make requests to the list of endpoints below. You can test if the server is running using:
`curl http://localhost:3333/getBalance`

### Endpoints 
The port used is 3333 by the way.

`http://localhost:3333/getBlockchainInfo` --> Returns the information of the blockchain as a string. 

`http://localhost:3333/getNewAddress` --> Creates a new recipient address for the currently running wallet. You can use this to create an address for mining rewards, or for a transaction, for example. 

`http://localhost:3333/getBalance` --> Gets the balance of the currently running wallet.

`http://localhost:3333/mine` --> Turns the background OrcaNet node into a mining node. Mining rewards will go to the associated wallet (the one running on your system)

`http://localhost:3333/sendToAddress` --> Takes a JSON object of the form:
```json
{ 
    "coins": "<num-coins>",
    "address": "<recipient-address>",
    "senderwalletpass": "<password to unlock wallet"
}
```
It sends `num-coins` to `recipient-address` by first using `senderwalletpass` to unlock the wallet

