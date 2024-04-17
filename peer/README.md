# Peer Node

## Basic Functionality

* Retrieving
    1) Find addresses corresponding with a specific file hash inside the DHT
    2) Connect through HTTP with the cheapest option
    3) Recieve chunks from the HTTP server
        1) If the file is small (< 4KB), the entire transaction is done in one go
        2) Otherwise, files are sent in chunks
            1) A signed transaction must be sent to the sender
    4) Close connection
    5) Add file to imported folder inside files directory

* Storing
    1) Hash the file content that you are trying to store
    2) Send a request to store the file in to the DHT
    3) Make sure file is inside the stored folder inside files directory
    4) Accept/Decline any requests for said file, or set the default behavior
    5) Send all microtransactions to blockchain once everything is done


## Assumptions

1) Each consumer/producer has their own IP address

2) Producer sets up local HTTP server

3) Consumer can fetch document from producer's local HTTP server

## Installation

* Information about installig proto buffer compiler is found [HERE](https://grpc.io/docs/protoc-installation/)

The basic steps are:

```bash

$ apt install -y protobuf-compiler

$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28

$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

$ export PATH="$PATH:$(go env GOPATH)/bin"


## Running

First generate the gRPC files for GO. Make sure you are in the root of the project and run the command below.

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

$ make

```

GO Version: 1.21.4

```bash

$ make all

```

## CLI interface

Get a file from the DHT.

```bash

$ get [fileHash] 

```

Storing a file in the DHT for a given price.

```bash
$ store [fileHash] [amount]
```

Import a file into the files directory:

```bash
$ import [filepath]
```

Send a certain amount of coin to an address

```bash

$ send [amount] [ip] 

```

Hash a file

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

1. Route /uploadFile is a POST route. This will move a local file, from anywhere on the computer, into the <i>files</i> directory. You should give the absolute path, or a path that is relative to the root directory of the peer node folder.

Request Body:
```json
{
    "filepath": "string"
}
```

---

2. Route /deleteFile is a POST route. This will delete a file that is stored from within the files folder

Request Body:
```json
{
    "filename": "string",
    "filepath": "string"
}
```

---

3. Route /getAllFiles is a GET route. This will return a json list of all files that are in the <i>files/</i> directory. This is a list of all files that have been imported by the user from the local machine.

Request Body: NONE

Response Body:
```json
[
    {
    "filename": "string",
    "filesize": "integer",
    "filehash": "string",
    "lastmodified":"string"
    },
]
```

---

4. Route /getAllStoredFiles is a GET route. This will return a json list of all files that are in the <i>files/stored</i> directory. This is a list of all files that are being stored by the peer on the network.

Request Body: NONE

Response Body:
```json
[
    {
    "filename": "string",
    "filesize": "integer",
    "filehash": "string",
    "lastmodified":"string"
    },
]
```

---

5. Route /getAllRequestedFiles is a GET route. This will return a json list of all files that are in the <i>files/requested</i> directory. THis a list of all the files requested by the peer.

Request Body: NONE

Response Body:
```json
[
    {
    "filename": "string",
    "filesize": "integer",
    "filehash": "string",
    "lastmodified":"string"
    },
]
```

---

6. Route /requestFile/:filename with a GET Request. You will need to pass the name of the file. This is called by the peer-node itself to handle file transfer.

Example: GET /requestFile/in.txt

Request Body: NONE

Response Body:
```json
{
    "status": "string"
}
```

---

7. Route /storeFile/:filename with a GET Request, similar to the route /requestFile. This is called by the peer-node itself to send market a notice that THIS peer-node is storing a file on this local machine.

Example: GET /requestFile/in.txt

Request Body: NONE

Response Body:
```json
{
    "status": "string"
}
```

---

8. Route /sendTransaction with a POST Request, must send the transaction and a signed version of the transaction. The body should be an octet-stream of the json object that is described below in the Request Body.

Request Body: 
```json
{
    "bytes": "bytes[]",
    "transaction": "byte[]",
    "public_key": "string"
}
```

Response Body:
```json
{
    "status": "string"
}
```

---

9. Route /getFileInfo?filename="" is a GET route. This will return the status of a file that was found in the files directory. The filecontent is a base64 string.

Request Body: NONE

Response Body:
```json
 {
    "filename": "string",
    "filesize": "integer",
    "filehash": "string",
    "lastmodified":"string",
    "filecontent":"string"
}
```

---

10. Route /writeFile is a POST route. This is the API route for the uploadfile function. It will take the contents of a file as a base64 string, write it to the specified file.

Request Body: 
```json
{
	"base64File":"string", 
	"fileSize":"string", 
	"originalFileName":"string" 
}
```

---

11. Route /updateActivityName is a POST Request. You will update the name of an activity byt providing a NEW name and the activity's id.

Request Body: 
```json
{
    "id":"int",
    "name":"string"
}
```

---

12. Route /removeActivity is a POST Request. It will remove a tracked activity that

Request Body: 
```json
{
    "id":"int",
}
```

---

13. Route /getActivities is a GET Request. It will return a list of all activities that are currently being tracked and stored.

Request Body: NONE

Response Body:
```json
[
    {
        "id":"int" ,  
        "name":"string",
        "size": "string",
        "hash": "string",
        "Status": "string",
        "showDropdown": "bool",  
        "peers": "int"    
    }
]
```

---

14. Route /setActivity is a POST Request. This will add a new activity to the list of activities that is being tracked.

Request Body:
```json
{
    "id":"int" ,  
    "name":"string",
    "size": "string",
    "hash": "string",
    "Status": "string",
    "showDropdown": "bool",  
    "peers": "int"    
}
```

---

15. Route /addPeer is a POST Request. It will add a new peer to the list of peers OUR peer is aware of.

Request Body:
```json
{
	"location":"string", 
	"latency":"string", 
	"peerID":"string", 
	"connection":"string", 
	"openStreams":"string", 
	"flagUrl":"string", 
}
```

---

16. Route /getPeer is a GET Request. It will get information about a peer, given a peer ID.

Request Body:
```json
{
	"peerID":"string",
}
```

Response Body:
```json
{
	"location":"string", 
	"latency":"string", 
	"peerID":"string", 
	"connection":"string", 
	"openStreams":"string", 
	"flagUrl":"string", 
}
```

---

17. Route /getAllPeers is a GET Request. It will get you an array of information about every peer node THIS peer node is aware of. 

Request Body: NONE

Response Body:
```json
[
    {
        "location":"string", 
        "latency":"string", 
        "peerID":"string", 
        "connection":"string", 
        "openStreams":"string", 
        "flagUrl":"string", 
    }
]
```

---

19. Route /updatePeer is a POST Request. It will update the an entire status of a peer, given a peer id. You cannot currently change the peer id from a POST Request. You would have to remove the peer and then add it back through different HTTP routes.

Request Body:

```json
{
	"location":"string", 
	"latency":"string", 
	"peerID":"string", 
	"connection":"string", 
	"openStreams":"string", 
	"flagUrl":"string", 
}
```

---

20. Route /removePeer is a POST Request. It will remove a peer based on a specific peerId.

Request Body:

```json
{
	"peerID":"string",
}
```

---

21. Route /sendMoney is a POST Request. It will attempt to send a signed transaction of a certain amount of money to a user. It will use the public and private key files that are stored inside the config folder.

Request Body:

```json 
{
    "amount":"float64", 
	"host":"string",  
	"port":"string"  
}
```

---

22. Route /getLocation is a GET Request. It will get the current location of THIS peer node.

Request Body: None

Response Body: 

```json
{
    "ip":        "string", 
	"network":   "string",
	"city":      "string",
	"region":    "string", 
	"country":   "string", 
	"latitude":  "string", 
	"longitude": "string", 
	"asn":       "string", 
	"timezone":  "string", 
	"continent": "string", 
	"org":       "string" 
}
```

---

23. Route /hash is a POST Request. It will return the hash of a file. It startes looking for files in the root of the files directory only. The file will need to be in the ./files/ directory in order to be found to be hashed. This can change if needed.

Request Body: 

```json
{
    "filepath": "string"
}
```

Response Body: 

```json
{
    "hash": "string"
}
```

---

24. Route /getBalance is a GET Request. It will retrieve the current running wallet balance

Request Body: NONE

Response Body:
```json
{
    "balance": "float64"
}
```

--- 

25. Route /walletPassphrase is a POST Request. It will unlock a wallet for a specified amount of time. walletName is the name of the wallet and timeUnlock is a string in milliseconds that specifies how long the wallet should stay unlocked for. NOTE : Current wallet needs to be unlocked before proceeding with money transfer

Request Body:
```json
{
    "walletName":"string",
    "timeUnlock":"string"
}
```

Response Body:
```json
{
    "status":"string"
}
```

---

26. Route /sendToAddressCommand is a POST Request. This will transfer Orca Coin from current wallet to the specified wallet address. You specify the destination wallet address and the amount of money. This route will return a hash of the transaction.

Request Body: 
```json
{
    "walletAddress":"string",
    "amount":"string"
}
```

Response Body:
```json
{
    "hash":"string"
}
```

---

27. Route /generateCommand is a POST Request. It will mine the specified number of blocks and sends rewards to the
running wallet. Specify the amount of blocks to mine and the block hashes will be returned in an array.

Request Body: 
```json
{
    "blocks":"string",
}
```

Response Body:
```json
[
    {
        "hash":"string"
    }
]
```






