## Records 
Our market records will be validated for the following specification:

```
                Array of Bytes

     +-----------------------------------+
     |User Protocol Buffer Message Length|
     |             (2 Bytes)             |
     +-----------------------------------+
     +-----------------------------------+
     |     Digital Signature Length      |
     |             (2 Bytes)             |
     +-----------------------------------+
     +-----------------------------------+
     |    User Protocol Buffer Message   |
     |             (Variable)            |
     +-----------------------------------+
     +-----------------------------------+
     |        Digital Signature          |
     |             (Variable)            |
     +-----------------------------------+
                       |
                  (Repeating)
                       |
                       v
     +-----------------------------------+
     |             UTC Time              |
     |             (8 Bytes)             |
     +-----------------------------------+
```

1) Each signature of the user protocol buffer message must be valid or the DHT will not accept the chain.
2) There can only be one record per public key in a chain or the DHT will not accept the chain.
3) The DHT will select values based on the latest, longest chain.