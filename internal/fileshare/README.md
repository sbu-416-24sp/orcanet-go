# Service Description

Further describing the purpose of each function and who should be the client & server

## rpc RecordFileStoreTransaction(TransactionStore) returns (TransactionACKResponse);

This would be sent by either the producer or consumer to the the market of a transaction so it can be stored and verified in the blockchain. This could go directly to the blockchain, but it might be important to let the market know that a request has been fulfilled and who fulfilled it. This is specifically for REQUESTING STORAGE OF ONE FILE.

## rpc RecordFileRequestTransaction(TransactionRequest) returns (TransactionACKResponse);

This would be sent by either the producer or consumer to the the market of a transaction so it can be stored and verified in the blockchain. This could go directly to the blockchain, but it might be important to let the market know that a request has been fulfilled and who fulfilled it. This is specifically for REQUESTING ONE SPECIFIC FILE.

## rpc PlaceFileStoreRequest(FileDesc) returns (StorageIP);

This would be sent from the consumer to the market to let the market know of the request for storage of a file. The market should resolve this request by sending back the IP of where a potential storage partner. 

## rpc PlaceFileRequest (FileDesc) returns (StorageIP);

This would be sent from the consumer to the market to let the market know of the request of the file. The market should then resolve this request by sending back the IP of where the file is stored.

## rpc NotifyAvailableStorage(StorageIP) returns (StorageACKResponse);

This would be sent from the producer to the market to let the market know there is space that is available for storage. The market just needs to acknowledge that it has received the request and is now looking to match an order to the producer so they can attempt to store a file.

## rpc NotifyStorageOppurtunity(stream FileDesc) returns (StorageResponse);

This would be sent from the market to the producer to let the producer know of a potential storage opportunity. The producer will need to acknowledge whether or not it wants to go ahead with this transaction or if it declines.

## rpc SendFile(FileDesc) returns (FileDesc);

This would send a file from the producer to the consumer.

## rpc SendFileToStore(FileDesc) returns (FileDesc);

This would send a file to store from the consumer to the producer. Not entirely sure if this usage of producer/consumer is correct here.