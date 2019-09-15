package datachain

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"text/template"
	"time"

	"ekyu.moe/cryptonight"
	_ "github.com/mattn/go-sqlite3"
)

type Block struct {
	Time     int64  `json:"t"`
	Nonce    int64  `json:"nonce"`
	Data     []byte `json:"data"`
	Prevhash []byte `json:"phash"`
}

type PrettyBlock struct {
	Time     time.Time `json:"time"`
	Nonce    int64     `json:"nonce"`
	Data     string    `json:"data"`
	Prevhash string    `json:"prev_hash"`
	Hash     string    `json:"hash"`
}

type Blockchain []Block

var (
	difficulty = 1
)

func LastBlocks(blockfile string, n int) Blockchain {
	out := Blockchain{}

	db, err := sql.Open("sqlite3", blockfile)

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM blockchain ORDER BY time,prevhash,nonce; DESC LIMIT %v", n))
	check(err)

	for rows.Next() {
		tmp := Block{}
		hash := []byte{}
		err = rows.Scan(tmp.Time, tmp.Nonce, tmp.Data, tmp.Prevhash, &hash)
		check(err)
		out = append(out, tmp)
	}
	rows.Close()
	db.Close()

	return out
}

//to-do: io.reader version for big chains
func (bc Blockchain) Tofile(blockfile string) {
	bc = bc.Comb()
	db, err := sql.Open("sqlite3", blockfile)
	check(err)
	for _, v := range bc {
		stmt, err := db.Prepare(`INSERT INTO blockchain
			(time, nonce, data, prevhash, hash) VALUES (?,?,?,?,?)
		`)
		check(err)
		data, _ := json.Marshal(v)
		hash := cryptonight.Sum(data, 0)
		_, err = stmt.Exec(v.Time, v.Nonce, v.Data, v.Prevhash, hash)
		check_and_skip(err)
	}
	// data, err := json.Marshal(bc)
	// check(err)
	// encdata := snappy.Encode(nil, data)
	// err = ioutil.WriteFile(blockfile, encdata, 0644)
	// check(err)
	db.Close()
}

func NewBlockchain(blockfile string) Blockchain {
	db, err := sql.Open("sqlite3", blockfile)

	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS blockchain(
		time BIG INT,
		nonce BIG INT,
		data TEXT,
		prevhash TEXT,
		hash TEXT PRIMARY KEY
	);`)

	check(err)

	_, err = stmt.Exec()

	// fmt.Printf("%+v", res)

	db.Close()
	return Blockchain{}
}

func BlockchainFromFile(blockfile string) Blockchain {
	// bc := Blockchain{}
	// data, err := ioutil.ReadFile(blockfile)
	// if err != nil {
	// 	return bc
	// }
	// encdata, err := snappy.Decode(nil, data)
	// check(err)
	// json.Unmarshal(encdata, &bc)
	// // fmt.Printf("blockchain: %+v", bc)
	// return bc
	db, err := sql.Open("sqlite3", blockfile)
	check(err)

	rows, err := db.Query("SELECT * FROM blockchain ORDER BY time,prevhash,nonce;")
	check(err)

	out := Blockchain{}

	for rows.Next() {
		tmp := Block{}
		hash := []byte{}
		err = rows.Scan(&tmp.Time, &tmp.Nonce, &tmp.Data, &tmp.Prevhash, &hash)
		check(err)
		out = append(out, tmp)
	}
	rows.Close()
	db.Close()
	return out
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
	b.Time = time.Now().UnixNano()

	data, err = json.Marshal(b)
	check(err)
	hash := cryptonight.Sum(data, 0)

	for countzero(hash) < difficulty {
		b.Nonce = rand.Int63()
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
		if !hashes[string(hash)] || bc[i-1].Time > bc[i].Time || bc[i].Time > time.Now().UnixNano() || len(bc[i].Prevhash) == 0 {
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
			bc = remove(bc, i)
			if i >= len(bc) {
				break
			}
			goto labelfor
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

func check_and_skip(e error) {
	if e != nil && fmt.Sprint(e)[:len("UNIQUE")] != "UNIQUE" {
		fmt.Fprintf(os.Stderr, "%+v\n", e)
	}
}

func queryBlocks(query, blockfile string) (Blockchain, error) {
	db, err := sql.Open("sqlite3", blockfile)
	// check(err)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query)
	// check(err)
	if err != nil {
		return nil, err
	}
	out := Blockchain{}

	for rows.Next() {
		tmp := Block{}
		hash := []byte{}
		err = rows.Scan(&tmp.Time, &tmp.Nonce, &tmp.Data, &tmp.Prevhash, &hash)
		if err != nil {
			return nil, err
		}
		// check(err)
		out = append(out, tmp)
	}
	rows.Close()
	db.Close()
	return out, nil
}

func (b Block) SaveBlock(blockfile string) error {
	db, err := sql.Open("sqlite3", blockfile)
	if err != nil {
		return err
	}
	stmt, err := db.Prepare(`INSERT INTO blockchain
			(time, nonce, data, prevhash, hash) VALUES (?,?,?,?,?)
		`)
	data, err := json.Marshal(b)
	hash := cryptonight.Sum(data, 0)
	stmt.Exec(b.Time, b.Nonce, b.Data, b.Prevhash, hash)
	db.Close()
	return nil
}

func (b Block) Prettify() PrettyBlock {
	data, err := json.Marshal(b)
	check(err)
	prev := string(hex.EncodeToString(b.Prevhash))
	curr := string(hex.EncodeToString(cryptonight.Sum(data, 0)))
	return PrettyBlock{time.Unix(0, b.Time), b.Nonce, string(b.Data), prev, curr}
}

type PrettyChain []PrettyBlock

func (bc Blockchain) Prettify() PrettyChain {
	out := make(PrettyChain, len(bc))
	for i := range bc {
		out[i] = bc[i].Prettify()
	}
	return out
}

func Server(path string, bc Blockchain, blockfile string) {
	bc = bc.Comb()
	bc.Tofile(blockfile)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("../index.html"))
		if r.Method != http.MethodPost {
			data := struct {
				Bc  PrettyChain
				Err bool
			}{Bc: bc.Prettify(), Err: false}
			tmpl.Execute(w, data)
			return
		}
		data := struct {
			Bc  PrettyChain
			Err bool
		}{Err: false}
		tmp, err := queryBlocks(r.FormValue("query"), blockfile)
		data.Bc = tmp.Prettify()
		if err != nil {
			// panic(err)
			data := struct {
				Err bool
				Msg error
			}{true, err}
			tmpl.Execute(w, data)
			return
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("../add.html"))
		if r.Method != http.MethodPost {
			tmpl.Execute(w, bc[len(bc)-1].Prettify())
			return
		}
		b := Block{Data: []byte(r.FormValue("data"))}
		b.Mine(bc[len(bc)-1])
		b.SaveBlock(blockfile)
		p := b.Prettify()
		tmpl.Execute(w, p)
		// http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Write([]byte("The request must be POST"))
			return
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte("Invalid body"))
			return
		}
		qry := struct {
			Query string
		}{}
		err = json.Unmarshal(data, &qry)
		if err != nil {
			w.Write([]byte("the body must be a JSON containing a field named Query, with the SQL query as a string"))
			return
		}
		bc, err := queryBlocks(qry.Query, blockfile)
		if err != nil {
			w.Write([]byte("Error on running query" + fmt.Sprint(err)))
		}
		out, err := json.Marshal(bc.Prettify())
		if err != nil {
			w.Write([]byte("Error on the blockchain" + fmt.Sprint(err)))
			return
		}
		w.Write(out)
	})

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(path, nil)
}
