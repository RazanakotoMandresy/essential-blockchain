package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Block struct {
	Pos       int
	Data      BookCheckout
	TimeStamp string
	Hash      string
	PrevHash  string
}

type BookCheckout struct {
	BookId       string `json:"book_id"`
	User         string `json:"user"`
	ChechoutDate string `json:"checkout_date"`
	IsGenesis    bool   `json:"is_genesis"`
}

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	PublishDate string `json:"publish_date"`
	ISBN        string `json:"isbn"`
}
type BlockChain struct {
	blocks []*Block
}

var Blockchain *BlockChain

func (b Block) generateHash() {
	bytes, _ := json.Marshal(b.Data)
	data := string(b.Pos) + b.TimeStamp + string(bytes) + b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

func CreateBlock(prevBlock *Block, checkoutitem BookCheckout) *Block {
	block := &Block{}
	block.Pos = prevBlock.Pos + 1
	block.TimeStamp = time.Now().String()
	block.PrevHash = prevBlock.Hash
	block.generateHash()
	return block
}
func ValidBlock(block, prevBlock *Block) bool {
	if prevBlock.Hash != block.PrevHash {
		return false
	}
	if !block.validateHash(block.Hash) {
		return false
	}
	if prevBlock.Pos+1 != block.Pos {
		return false
	}
	return true
}
func (b *Block) validateHash(hash string) bool {
	b.generateHash()
	if b.Hash != hash {
		return false
	}
	return true
}
func (b *BlockChain) Addblock(data BookCheckout) {
	prevBlock := b.blocks[len(b.blocks)-1]
	block := CreateBlock(prevBlock, data)
	if ValidBlock(block, prevBlock) {
		b.blocks = append(b.blocks, block)
	}
}
func newBook(w http.ResponseWriter, r *http.Request) {
	book := Book{}
	if err := json.NewDecoder(r.Body).Decode(book); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("can't create a book %v", err)
		w.Write([]byte("can't create a book"))
		return
	}
	h := md5.New()
	io.WriteString(h, book.ISBN+book.PublishDate)
	book.ID = fmt.Sprint(h.Sum(nil))
	res, err := json.MarshalIndent(book, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf(" could not marshal %v", err)
		w.Write([]byte("could not save book data"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkoutItem BookCheckout
	if err := json.NewDecoder(r.Body).Decode(&checkoutItem); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not write the block:%v", err)
		w.Write([]byte("could not write block"))
	}
	Blockchain.Addblock(checkoutItem)
}
func GenesisBlock() *Block {
	return CreateBlock(&Block{}, BookCheckout{IsGenesis: true})
}
func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{GenesisBlock()}}
}
func getBlock(w http.ResponseWriter, r *http.Request) {
	jbyte, err := json.MarshalIndent(Blockchain.blocks, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	io.WriteString(w, string(jbyte))

}
func main() {
	Blockchain = NewBlockChain()
	port := os.Getenv("PORT")
	r := mux.NewRouter()
	r.HandleFunc("/", getBlock).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newBook).Methods("POST")
	go func() {
		for _, block := range Blockchain.blocks {
			fmt.Printf("Prev. hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data:%v\n", string(bytes))
			fmt.Printf("hash:%x\n", block.Hash)

		}
	}()
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err, "error when listening and serve the server")
	}
	log.Println("listing and serve on ", port)
}
