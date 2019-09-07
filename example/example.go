package main

import (
	bc "github.com/cauefcr/datachain"
	"fmt"
	"time"
)


func main() {
	var (
		blockchain = bc.Blockchain{}
		blockfile  = "./blockchain.json"
	)
	blockchain = bc.BlockchainFromFile(blockfile)
	defer blockchain.Tofile(blockfile)
	nb := bc.Block{Time: uint64(time.Now().UnixNano()), Data:[]byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")}
	blockchain = append(blockchain,nb)
	nb.Mine(nb)
	blockchain = append(blockchain,nb)
	nb.Mine(nb)
	// blockchain.Mine([]byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"))
	// fmt.Printf("%+v", blockchain)
	// uncomment to have an orphaned block
	blockchain = append(blockchain, bc.Block{Data: []byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")})
	if clean_chain := blockchain.Comb(); len(clean_chain) != len(blockchain) {
		fmt.Println("blocks were orphaned")
		blockchain = clean_chain
	} else {
		fmt.Println("No orphans, the blockchain was clean")
	}
	fmt.Printf("%+v",blockchain)
}
