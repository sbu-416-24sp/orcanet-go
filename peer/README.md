# Peer Node

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