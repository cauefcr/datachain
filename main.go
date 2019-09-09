package datachain

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"time"

	// "fmt"
	"ekyu.moe/cryptonight"
	"github.com/golang/snappy"
)

type Block struct {
	Time     uint64 `json:"t"`
	Nonce    uint64 `json:"nonce"`
	Data     []byte `json:"data"`
	Prevhash []byte `json:"phash"`
}

type Blockchain []Block

var (
	difficulty = 1
)

//to-do: io.reader version for big chains
func (bc Blockchain) Tofile(blockfile string) {
	data, err := json.Marshal(bc)
	check(err)
	encdata := snappy.Encode(nil, data)
	err = ioutil.WriteFile(blockfile, encdata, 0644)
	check(err)
}

func BlockchainFromFile(blockfile string) Blockchain {
	bc := Blockchain{}
	data, err := ioutil.ReadFile(blockfile)
	if err != nil {
		return bc
	}
	encdata, err := snappy.Decode(nil, data)
	check(err)
	json.Unmarshal(encdata, &bc)
	// fmt.Printf("blockchain: %+v", bc)
	return bc
}

func (b *Block) Mine(prev Block) {
	// var prev Block
	countzero := func(h []byte) int {
		i := 0
		for ; i < len(h) && h[i] == 0; i++ {

		}
		return i
	}
	data, err := json.Marshal(prev)
	check(err)
	b.Prevhash = cryptonight.Sum(data, 0)
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
}

func remove(slice []Block, s int) []Block {
	return append(slice[:s], slice[s+1:]...)
}

func (bc Blockchain) Comb() Blockchain {
	if len(bc) == 0 {
		return bc
	}
	i := 0
	hashes := map[string]bool{}
	data, err := json.Marshal(bc[i])
	check(err)
	hash := cryptonight.Sum(data, 0)
	hashes[string(hash)] = true
	// j := 0
	for i = 1; i < len(bc); i++ {
		hashes[string(hash)] = true
	labelfor:
		if !hashes[string(hash)] || bc[i-1].Time > bc[i].Time || bc[i].Time > uint64(time.Now().UnixNano()) || len(bc[i].Prevhash) == 0 {
			// fmt.Printf("%v(len(%v),cap(%v)),", i, len(bc), cap(bc))
			// break
			// j++
			// hashes[string(hash)] = false
			bc = remove(bc, i)
			if i >= len(bc) {
				break
			}
			goto labelfor
			// i--
		}
		data, err := json.Marshal(bc[i])
		if err != nil {
			// hashes[string(hash)] = false
			// remove(bc, i)
			// i--
			// break
		}
		hash = cryptonight.Sum(data, 0)
	}
	// bc = bc[:i]

	return bc
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
