# Peer Node

## Basic Functionality

1) Request from the market the ip of a file

2) Send request to that specific ip

3) Record transaction with producer to blockchain

#### Notes:

* Files that are on the network should be in the files folder. This can be done manually or by using the CLI

* Inside the config file, set your public key and private key location from the 


## Assumptions

1) Each consumer/producer has their own IP address

2) Producer sets up local HTTP server

3) Consumer can fetch document from producer's local HTTP server


## Running

First generate the gRPC files for GO. Make sure you are in the root of the project and run the command below.

``` bash

$ protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/fileshare/file_share.proto 
```

GO Version: 1.21.4

```bash

$ make all

```

## CLI interface

Requesting a file:

```bash
$ get [ip] [port] [filename]
```

Storing a file:

```bash
$ store [ip] [address] [filename]
```

Import a file:

```bash
$ import [filepath]
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


## gRPC API

* RecordFileRequestTransaction: Tell blockchain of a completed transaction

* RequestAllAvailableFileNames

* PlaceFileRequest: Ask market to tell you ALL possible locations where file is store 

* NotifyFileStore Tell market you will store a file for future access

* NotifyFileUnstore: Tell market you no longer have a specific file

* SendFile: Send a File

## HTTP Functionality

Server should only look for two things:

* Route /requestFile with a GET Request, parameter of `filename`, a string that represents name of file

* Route /storeFile with a GET Request, similar to the route /requestFile





