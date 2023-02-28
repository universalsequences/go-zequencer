# go-zequencer

An efficient (in-memory) cache server of Ethereum Event Logs. Allows for constructing SQL-like search queries on "tables" (Ethereum Event Types), 
partitioned by contract.

To compile, simply do:

go build

To run:

./go-zequencer *path-to-directory-of-logs*

# How it works

go-zequencer reads every file from a directory that contains files named by the contract they contain events for.
Each file is a json where each field is an ethereum event type:

Ex: CONTRACT_ADDRESS_1.json
{
   eventType1: [{field1: 100, field2: 200, blockNumber: 20403}, {field1: 120, field2: 230, blockNumber: 20444}]
   eventType2: ... etc
}

# How I Use It

1. I have a seperate node.js script that fetches all the events for a set of (contract, eventType) pairs and writes the outputs to these files.
2. When any event occurs & updates the files, I sends an http request to the go-zequencer server saying "we've updated contract x"
3. go-zequencer reads the files for "contract x" (that has just updated), and updates the in-memory mapping of the entire event space

# Limitations
This is bad for when you have a ton of data (like storing every NFT ever). However, will perform significantly faster than GraphQL 
for "smaller/medium" amounts of data. 

The main usecase here, is when you care about a certain set of contracts and its events, and want lightning fast data availability.
