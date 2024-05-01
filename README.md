# orcanet-go

Main Repo for Orcanet-Go 🐳

Orcanet is a peer-to-peer file-sharing service. The project allows users to pay other users to store their files, as well as to request files. Payments are made using cryptocurrency. For this project, we implemented our own testnet, which is built off of BTCD. 

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


# Running the Peer Node

To get started running the peer node, navigate into the peer directory:
`cd peer`

## Installation

* Make sure golang is installed: The version used for this project is 1.21.4
* Information about installing proto buffer compiler is found [HERE](https://grpc.io/docs/protoc-installation/)

The basic steps are:

```bash

$ apt install -y protobuf-compiler

$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28

$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

$ export PATH="$PATH:$(go env GOPATH)/bin"


## Running

First generate the gRPC files for GO. Make sure you are in the peer folder of the project and run the command below.

``` bash

$ protoc --go_out=./internal/fileshare \
    --go_opt=paths=source_relative \
    --go-grpc_out=./internal/fileshare \
    --go-grpc_opt=paths=source_relative \
    --proto_path=../protos/fileshare \
    file_share.proto
```

Second, make sure you create the executable for the bitcoin nodes. It can be done from the root directory as follows:

```bash

$ cd coin

$ ./build.sh

```

Finally, return the the peer folder and run make.

```bash

$ cd ../peer

$ make all

```

This will start up the peer node. You should see output in the terminal. You will need to enter in <i>three</i> numbers into the terminal before the peer node is fully running. These three numbers will be the port numbers used by the peer node to connect with various services. There is no agrred upon port number, but currently, these three ports can be the official

## CLI interface

Get a file from the DHT. You should pass a specific hash.

```bash

$ get [fileHash] 

```

Storing a file in the DHT for a given price. You should pass ONLY the file name, given the file is in the files folder (inside peers).

```bash
$ store [filename] [amount]
```

Import a file into the files directory. You can pass it any filepath, but if the path is relative. It will be rooted in the ./peer folder. It is best to just use an absolute path.

```bash
$ import [filepath]
```

Send a certain amount of coin to an address

```bash

$ send [amount] [ip] 

```

Hash a file. Only files inside the files folder can be found. Only pass relative paths. You should not need to hash any files: this should be handled internally.

```bash

$ hash [filename]

```

Listing all files stored for IPFS

```bash
$ list
```

Getting current peer node location

```bash
$ location
```

Testing network speeds

```bash
$ network
```

Exiting Program

```bash
$ exit
```

#### File System:

* There is a folder called <i>files</i>. This is where all the files that are available to the user is stored

* Any file directly stored inside <i>files</i> folder is considered <i>uploaded</i> to the client.

* Any file that has been requested by the user is stored in the <i>files/requested</i> folder.

* Any file that is available to be requested for by anyone on the network is in <i>files/stored</i>.

* Technically, you can import the files manually if you drag them inside the desired folder. There is currently no protection against this.

* The <i>transactions</i> folder stores all of the transactions that have been processed and stored.

#### Notes:

* Files that are on the network should be in the files folder. This can be done manually or by using the CLI

* Inside the config file, set your public key and private key location. If you don't want to, the CLI will generate a key-pair for you.

* Only .txt, .json and .mp4 file formats are currently supported.


## HTTP Functionality

Here is all of the routes available on the HTTP server that is started when the peer-node loads. Most routes should return 400 if an issue with the parameters sent by the client did not work, 405 if the wrong method type (GET, POST) was used, 500 if there was an error creating, searching or opening files and 200 if everything is successful. If a response is sent inside of an array, that indicates that at minimum, 0 json objects could be sent but more than 1 json object could also be inside of that json array. If not explicitly stated, the Response Body should be a json object with a single field name "status", explaing the current status of the request. Furthermore, when any response code other than 200 is sent, there should be this same json object sent inside the Response Body.

---

Routes should follow the API laid out in the document from the front end team. 


Some additional internal routes we added for communicating between peer nodes are below. These should be peer to peer only, not front-end to peer.

/requestFile/
/sendTransaction
/writeFile
/sendMoney
/getLocation
/getAllStored
/get-file
/upload-file
/delete-file

The blockchain routes that currently exist are as follows. We still need to fix it to match the specification.

/getBlockchainInfo
/getNewAddress
/getBalance
/mine
/sendToAddress

Settings have not been implmented. Are we keeping it on the front-end?

Statistics are also a work in progress.

All other routes, should be as follows on the API document.


## TODO
### Features that should be implemented in future pull requests
1. **Team Sea Dolphins DHT Bad Address Connection** 
    - Implement the Sea Dolphins method of trying to reconnect to a peer on a bad address 3 times and then removing it from the peer's address book. 
2. **NAT Address Translation** 
    - Right now the peer node will join the DHT in client mode by default, which will only allow it to send out queries and not respond to them. The desired functionality is to join the DHT as a server node automatically if it can be determined that we can reach the node behind the NAT.
3. **NAT File Request**
    - Trying to store a file on a host that is behind a NAT will lead to an IO timeout. The peer will attempt to retrieve such a file, but will be unsuccessful. The peer can store a file on an address behind a NAT, whether or not this is allowed needs to be determined.
