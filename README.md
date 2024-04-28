# orcanet-go

Main Repo for Orcanet-Go 🐳

## TODO
### Features that should be implemented in future pull requests
1. **Team Sea Dolphins DHT Bad Address Connection** 
    - Implement the Sea Dolphins method of trying to reconnect to a peer on a bad address 3 times and then removing it from the peer's address book. 
2. **NAT Address Translation** 
    - Right now the peer node will join the DHT in client mode by default, which will only allow it to send out queries and not respond to them. The desired functionality is to join the DHT as a server node automatically if it can be determined that we can reach the node behind the NAT.
3. **NAT File Request**
    - Trying to store a file on a host that is behind a NAT will lead to an IO timeout. The peer will attempt to retrieve such a file, but will be unsuccessful. The peer can store a file on an address behind a NAT, whether or not this is allowed needs to be determined.