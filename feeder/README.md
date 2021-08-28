## Feeder

### Usage

$ Feeder \<kafka addr> \<num of messages> \<num of txs per message> \<len per tx>

### Example

$ Feeder "localhost:9092" 100000 1000 100

> Send 100000 messages to localhost:9092, every message contains 1000 transactions, all the transactions are 100 bytes. 