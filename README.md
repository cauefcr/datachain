# datachain
A simple (DAG) blockchain that can hold arbitrary data, uses CryptoNight hash function.

## Usage

```go
package main

import (
	"fmt"

	bc "github.com/cauefcr/datachain"
)

func main() {
	var (
		blockchain = bc.Blockchain{}
		blockfile  = "./blockchain.json"
	)
	// load blockchain from file (if it exists)
	blockchain = bc.BlockchainFromFile(blockfile)
	// make a new block
	nb := bc.Block{Data: []byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")}

	// mine the block
	nb.Mine(nb)
	// put it in the blockchain
	blockchain = append(blockchain, nb)

	fmt.Printf("%+v", blockchain)

	// uncomment to have an orphaned block
	// blockchain = append(blockchain, bc.Block{Data: []byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")})
	// check to see if the blockchain data makes sense, it will remove everything after a broken node
	if clean_chain := blockchain.Comb(); len(clean_chain) != len(blockchain) {
		fmt.Println("blocks were orphaned")
		blockchain = clean_chain
	} else {
		fmt.Println("No orphans, the blockchain was clean")
	}
	//view blockchain
	fmt.Printf("%+v", blockchain)

	// save blockchain to file
	blockchain.Tofile(blockfile)
}

```

## Features:

[x]Mine

[x]Saving

[x]Reading

[ ]Synching with other nodes

[x]Dealing with tree-like structures that can arrive

[ ]Querying structural data about the DAG

[ ]Subscribing to sub-tree add events
