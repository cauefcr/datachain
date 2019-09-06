package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"ekyu.moe/cryptonight"
)

//todo:
//sync blockchain with other nodes with some p2p thing
//make blockchain create, comb, append, get

type Block struct {
	Time     uint64 `json:"t"`
	ID       uint64 `json:"n"`
	Nonce    uint64 `json:"nonce"`
	Data     []byte `json:"data"`
	Prevhash []byte `json:"phash"`
}

var (
	blockchain = []Block{}
	blockfile  = "./blockchain.json"
	difficulty = 1
)

func (b *Block) Mine() {
	var prev Block
	countzero := func(h []byte) int {
		i := 0
		for ; i < len(h) && h[i] == 0; i++ {
			// fmt.Print(h[i])
		}
		if i != 0 {
			fmt.Println(i)
		}
		return i
	}
	if len(blockchain) > 0 {
		prev = blockchain[len(blockchain)-1]
	} else { // may satoshi have his keys
		prev = Block{ID: 0, Data: []byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"), Prevhash: []byte("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"), Time: uint64(time.Now().UnixNano())}
		hash := []byte{}
		for countzero(hash) < difficulty {
			prev.Nonce = rand.Uint64()
			data, _ := json.Marshal(prev)
			// check(err)
			hash = cryptonight.Sum(data, 0)
		}
		blockchain = append(blockchain, prev)
	}
	data, err := json.Marshal(prev)
	check(err)
	b.Prevhash = cryptonight.Sum(data, 0)
	b.ID = prev.ID + 1
	b.Time = uint64(time.Now().UnixNano())

	data, err = json.Marshal(b)
	check(err)
	hash := cryptonight.Sum(data, 0)

	for countzero(hash) < difficulty {
		b.Nonce = rand.Uint64()
		data, err = json.Marshal(b)
		check(err)
		hash = cryptonight.Sum(data, 0)
	}
	blockchain = append(blockchain, *b)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func init() {
	//check existence of the blockchain file
	//load it
	//check for it's correctness
	//if error fork on block
	//capture os signals to dump blockchain to disk
}

func main() {
	b := Block{Data: []byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")}
	b.Mine()
	fmt.Printf("%+v", blockchain)
}
