package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	MINING_LEVEL = 5
	MINING_SENDER_ADDRESS = "Mining Payer Address"
	MINING_REWARD = 0.1
)
type Block struct {
	nonce int
	timestamp int64
	previousBlock [32]byte
	transactions []*Transaction
}

type BlockChain struct {
	
	chain []*Block
	minerBlockChainAddress string
}

type Transaction struct {
	amount float32
	sender string
	receiver string
}

func createBlock(previousBlock [32]byte,nonce int,transactions []*Transaction) *Block {
	block := new(Block)
	block.previousBlock = previousBlock
	block.nonce = nonce
	block.timestamp = time.Now().UnixNano()
	block.transactions = transactions
	return block

}
func (block *Block) toByteArray() []byte {
	bytes, _ := json.Marshal(struct {
		Nonce int					`json:"nonce"`
		Timestamp int64				`json:"timestamp"`
		PreviousBlock [32]byte		`json:"previous_block"`
		Transactions []*Transaction	`json:"transactions"`
	}{	Timestamp: block.timestamp,
		Nonce: block.nonce,
		PreviousBlock: block.previousBlock,
		Transactions: block.transactions,
	})
	return bytes
}

func (block *Block) getHash() [32]byte {
	return sha256.Sum256(block.toByteArray())
}

func (block *Block) toString()  {
	fmt.Printf("Nonce	%d\n",block.nonce)
	fmt.Printf("Timestamp	%d\n",block.timestamp)
	fmt.Printf("Previous Block	%x\n",block.previousBlock)
	fmt.Printf("%sTransactions%s\n",strings.Repeat("-",5),strings.Repeat("-",5))
	for _,transaction := range block.transactions{
		transaction.toString()
	}
}

func (blockChain *BlockChain) addBlock(previousBlock [32]byte,nonce int,transactions []*Transaction) *Block {

	block := createBlock(previousBlock,nonce,transactions)
	blockChain.chain = append(blockChain.chain,block)
	return block
}

func (blockChain *BlockChain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction ,0)
	for _,t := range blockChain.getLastBlock().transactions {
		transactions = append(transactions,
			createTransaction(t.sender,t.receiver,t.amount))
	}
	return transactions
}

func (blockChain *BlockChain) ValidProof(nonce int, lastBlock [32]byte, transactions []*Transaction, level int) bool {
	 zeroes := strings.Repeat("0",level)
	 guessBlock := Block{
	 	timestamp: 0,
	 	nonce:nonce,
	 	previousBlock: lastBlock,
	 	transactions: transactions,
	 }
	 guessHash := fmt.Sprintf("%x",guessBlock.getHash())
	 return guessHash[:level] == zeroes
}

func (blockChain *BlockChain) Mining() bool {
	transaction := createTransaction(MINING_SENDER_ADDRESS,blockChain.minerBlockChainAddress,MINING_REWARD)
	var transactions []*Transaction
	transactions = append(transactions,transaction)
	nonce := blockChain.ProofOfWork()
	blockChain.addBlock(blockChain.getLastBlock().getHash(),nonce,transactions)
	log.Println("action=mining, status=success")
	return true
}

//ProofOfWork returns the nonce value for the block
func (blockChain *BlockChain) ProofOfWork() int {

	   transactions := blockChain.CopyTransactionPool()
	   previousHash := blockChain.getLastBlock().getHash()
	   nonce := 0
	   for !blockChain.ValidProof(nonce,previousHash,transactions,MINING_LEVEL){
	   		nonce+=1
	   }
	   return nonce
}

func createBlockChain(minerBlockChainAddress string) *BlockChain  {
	block := &Block{}
	blockChain := new(BlockChain)
	blockChain.minerBlockChainAddress = minerBlockChainAddress
	blockChain.addBlock(block.getHash(),0,nil)
	return blockChain
}

func (blockChain *BlockChain) toString()  {
	for i, block := range blockChain.chain {
		fmt.Printf("%sBlock-%d%s\n",strings.Repeat("=",10),i,strings.Repeat("=",10))
		block.toString()
	}
}

func (blockChain *BlockChain) getLastBlock() *Block {
	return blockChain.chain[len(blockChain.chain)-1]
}

func (transaction *Transaction) toByteArray() []byte {
	bytes, _ := json.Marshal(struct {
		Sender string        `json:"sender"`
		Receiver string `json:"receiver"`
		Amount float32   `json:"amount"`
	}{	Sender: transaction.sender,
		Receiver: transaction.receiver,
		Amount: transaction.amount,
	})
	return bytes
}

func createTransaction(sender string, receiver string, amount float32) *Transaction {
	
	transaction := new(Transaction)
	transaction.sender = sender
	transaction.receiver = receiver
	transaction.amount = amount
	return transaction
}

func (blockChain *BlockChain) calculateTotalAmount(blockChainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, block := range blockChain.chain  {
		for _, transaction := range block.transactions {
		  if blockChainAddress==transaction.receiver {
		  	totalAmount+=transaction.amount
		  }
		  if blockChainAddress == transaction.sender {
		  	totalAmount-=transaction.amount
		  }
		}
	}
	return totalAmount
}
func (transaction *Transaction) toString() {
	fmt.Printf("Sender	%s\n",transaction.sender)
	fmt.Printf("Receiver	%s\n",transaction.receiver)
	fmt.Printf("Amount	%.1f\n",transaction.amount)
}

func main()  {
	log.SetPrefix("BlockUsingGo")
	minerBlockChainAddress := "BlockChain Miner Address"
	blockChain := createBlockChain(minerBlockChainAddress)
	fmt.Println(blockChain.getLastBlock().previousBlock)
	var transactions []*Transaction

	transaction := createTransaction("abc","xyz",1.32)
	transactions = append(transactions,transaction)
	nonce := blockChain.ProofOfWork()
	blockChain.addBlock(blockChain.getLastBlock().getHash(),nonce,transactions)
	blockChain.Mining()

	transactions = nil

	transaction = createTransaction("xyz","pqr",5.12)
	transactions = append(transactions,transaction)
	nonce = blockChain.ProofOfWork()
	blockChain.addBlock(blockChain.getLastBlock().getHash(),nonce,transactions)
	blockChain.Mining()

	blockChain.toString()

	fmt.Println("\nTotal Value for miner address is ",
		blockChain.calculateTotalAmount(minerBlockChainAddress))
}
