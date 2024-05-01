# API Server wrapper for OrcaNet(full node) and the OrcaWallet, and Btcctl 

To get started with running btcd and btcwallet, navigate into the coin directory:
`cd coin`

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

`http://localhost:3333/getPeerInfo` --> Returns information of all peers connected to this node (ip, last send, last receive, etc...) in JSON format

```json
  {
    "id": ,
    "addr": ,
    "addrlocal": ,
    "services": ,
    "relaytxes": ,
    "lastsend": ,
    "lastrecv": ,
    "bytessent": ,
    "bytesrecv": ,
    "conntime": ,
    "timeoffset": ,
    "pingtime": ,
    "version": ,
    "subver": ,
    "inbound": ,
    "startingheight": ,
    "currentheight": ,
    "banscore": ,
    "feefilter": ,
    "syncnode":
  }
```

`http://localhost:3333/getBestBlock` --> Returns the hash and block height number of the best (most recently mined) block in JSON format. 

```json
{
    "hash":,
    "height":,
}```

`http://localhost:3333/getBestBlockInfo` --> Returns the information of the best (most recently mined) block in JSON format.

```json
{
  "hash": "00000000d003d6d26d3d51c6b0e39180c9ffe69386be33dc4e2f9eaeb914f458",
  "confirmations": 1,
  "strippedsize": 189,
  "size": 189,
  "weight": 756,
  "height": 99,
  "version": 536870912,
  "versionHex": "20000000",
  "merkleroot": "89a5f8fdb15df6a4a0545503d4963f8801ea7d771cdca29a33dd9cf78218ed59",
  "tx": [
    "89a5f8fdb15df6a4a0545503d4963f8801ea7d771cdca29a33dd9cf78218ed59"
  ],
  "time": 1714508126,
  "nonce": 1066119017,
  "bits": "1d00ffff",
  "difficulty": 1,
  "previousblockhash": "000000001d0bd78d22bf186e19ecc68ce72d5f227d0268654a1005abc15081bf"
}
```

`http://localhost:3333/stopMine` --> Restarts the OrcaNet node to a non mining instance. 


