package blockchain

import (
	"blockchaingo/utils"
	"blockchaingo/wallet"
	bytes2 "bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	MINING_LEVEL = 5
	MINING_SENDER_ADDRESS = "Mining Payer Address"
	MINING_REWARD = 0.1
	INTERNAL_TRANSACTION_API = "http://%s/internal/transaction"
)
type Block struct {
	nonce 					int						`json:"nonce,omitempty"`
	timestamp            	int64					`json:"timestamp,omitempty"`
	previousBlockAddress 	[32]byte				`json:"previous_block_address,omitempty"`
	Transactions         	[]*utils.Transaction	`json:"transactions,omitempty"`
}

type BlockChain struct {
	
	chain					[]*Block	`json:"chain"`
	minerBlockChainAddress 	string		`json:"miner_block_chain_address"`
	port 					uint16		`json:"port"`
	mux						sync.Mutex	`json:"mux"`
	muxNeighbors			sync.Mutex	`json:"mux_neighbors"`
	neighbors				[]string	`json:"neighbors"`
}

func (blockChain *BlockChain) SetNeighbors() {

	blockChain.neighbors = utils.FindBlockchainNeighbors(utils.GetHost(),blockChain.port,
		0,3,5000,5003)
	log.Printf("Blockchain %v neighbors %v",blockChain.minerBlockChainAddress,blockChain.neighbors)
	
}

func (blockChain *BlockChain) updateNeighbors() {
	
	blockChain.muxNeighbors.Lock()
	defer blockChain.muxNeighbors.Unlock()
	blockChain.SetNeighbors()
	
}

func (blockChain *BlockChain) CronUpdateNeighbors() {
	blockChain.updateNeighbors()
	_ = time.AfterFunc(time.Second*15,blockChain.CronUpdateNeighbors)
}

func (blockChain *BlockChain) CronUpdateBlockchainNetwork() {
	blockChain.UpdateBlockchainNetwork()
	_ = time.AfterFunc(time.Second*15,blockChain.CronUpdateBlockchainNetwork)
}

func (blockChain *BlockChain) Run() {
	 blockChain.CronUpdateNeighbors()
	 blockChain.CronUpdateBlockchainNetwork()
}

func (blockchain *BlockChain) Chain() []*Block {
	return blockchain.chain
}

func (blockchain *BlockChain) MinerBlockChainAddress() string {
	return blockchain.minerBlockChainAddress
}

func (blockchain *BlockChain) MarshalJSON() ([]byte,error) {
	return json.Marshal(struct {
			Blocks	[]*Block	`json:"chain"`
		}{
			Blocks: blockchain.chain,
		})
}

func (blockchain *BlockChain) UnmarshalJSON(bytes []byte) error {
	strct := &struct {
		Blocks	*[]*Block	`json:"chain"`
	}{
		Blocks: &blockchain.chain,
	}
	if err := json.Unmarshal(bytes, &strct); err!=nil{
		return err
	}
	return nil
}

func createBlock(previousBlock [32]byte,nonce int,transactions []*utils.Transaction) *Block {
	block := new(Block)
	block.previousBlockAddress = previousBlock
	block.nonce = nonce
	block.timestamp = time.Now().UnixNano()
	block.Transactions = transactions
	return block

}

func (block *Block) MarshalJSON() ([]byte,error) {
	return json.Marshal(struct {
		Nonce int					`json:"nonce"`
		Timestamp int64				`json:"timestamp"`
		PreviousBlockAddress string		`json:"previous_block_address"`
		Transactions []*utils.Transaction	`json:"transactions"`
	}{	Timestamp: block.timestamp,
		Nonce: block.nonce,
		PreviousBlockAddress: fmt.Sprintf("%x",block.previousBlockAddress),
		Transactions: block.Transactions,
	})
}

func (block *Block) UnmarshalJSON(bytes []byte) error {

	var prevHash string
	obj := &struct {
		Timestamp *int64	`json:"timestamp,omitempty"`
		Nonce	*int	`json:"nonce,omitempty"`
		PreviousBlockAddress	*string	`json:"previous_block_address,omitempty"`
		Transactions *[]*utils.Transaction	`json:"transactions,omitempty"`

	}{
		Timestamp: &block.timestamp,
		Nonce: &block.nonce,
		PreviousBlockAddress: &prevHash,
		Transactions: &block.Transactions,
	}
	if err := json.Unmarshal(bytes, &obj); err!=nil{
		return err
	}
	log.Println(*obj.PreviousBlockAddress)
	decodeStr,_ := hex.DecodeString(*obj.PreviousBlockAddress)
	copy(block.previousBlockAddress[:],decodeStr[:])
	return nil
}

func (block *Block) getHash() [32]byte {
	bytes, _ :=    block.MarshalJSON()
	return sha256.Sum256(bytes)
}

func (block *Block) toString()  {
	fmt.Printf("Nonce	%d\n",block.nonce)
	fmt.Printf("Timestamp	%d\n",block.timestamp)
	fmt.Printf("Previous Block	%x\n",block.previousBlockAddress)
	fmt.Printf("%sTransactions%s\n",strings.Repeat("-",5),strings.Repeat("-",5))
	for _,transaction := range block.Transactions {
		transaction.ToString()
	}
}

func (blockChain *BlockChain) addBlock(previousBlock [32]byte,nonce int,transactions []*utils.Transaction) *Block {

	block := createBlock(previousBlock,nonce,transactions)
	blockChain.chain = append(blockChain.chain,block)
	return block
}

func (blockChain *BlockChain) CopyTransactionPool(wallet *wallet.Wallet) []*utils.Transaction {
	transactions := make([]*utils.Transaction ,0)
	for _,t := range blockChain.GetLastBlock().Transactions {
		transactions = append(transactions,
			utils.CreateTransaction(wallet.PrivateKey(),
				wallet.PublicKey(),t.SenderBlockchainAddress(),t.RecipientBlockchainAddress(),
				t.Amount()))
	}
	return transactions
}

func (blockChain *BlockChain) ValidProof(nonce int, lastBlock [32]byte, transactions []*utils.Transaction, level int) bool {
	 zeroes := strings.Repeat("0",level)
	 guessBlock := Block{
	 	timestamp:            0,
	 	nonce:                nonce,
	 	previousBlockAddress: lastBlock,
	 	Transactions:         transactions,
	 }
	 guessHash := fmt.Sprintf("%x",guessBlock.getHash())
	 return guessHash[:level] == zeroes
}

func (blockChain *BlockChain) Mining(walet *wallet.Wallet) bool {

	//acquiring a mutex var to avoid parallel execution of mining since its a time intensive process
	log.Println("Mining Started")
	blockChain.mux.Lock()
	defer blockChain.mux.Unlock()
	if len(blockChain.Chain()) >1 &&
		blockChain.GetLastBlock().Transactions[0].SenderBlockchainAddress()==MINING_SENDER_ADDRESS{
		log.Println("Mining not required")
		return false
	}

	var transaction *utils.Transaction
	if walet ==nil{
		// creating the walet object from the last block
		log.Println("Creating Walet object for mining")
		walet =  wallet.CreateWalletWithKeys(
			blockChain.GetLastBlock().Transactions[0].SenderPrivateKey(),
			blockChain.GetLastBlock().Transactions[0].SenderPublicKey())
	}
	transaction = utils.CreateTransaction(  walet.PrivateKey(),walet.PublicKey(),
		MINING_SENDER_ADDRESS,blockChain.minerBlockChainAddress,MINING_REWARD)
	
	var transactions []*utils.Transaction
	transactions = append(transactions,transaction)
	
	nonce := blockChain.ProofOfWork(walet)
	blockChain.addBlock(blockChain.GetLastBlock().getHash(),nonce,transactions)


	// Need to update this info also to the blockchain neighbours
	for _, neighbor := range blockChain.neighbors{
		status := blockChain.sendTransactionToBlockchainNode(walet, transaction, neighbor)

		if !status {
			//Need to handle gracefully
			break
		}
	}

	log.Println("action=mining, status=success")
	return true
}

func (blockChain *BlockChain) CronMining() {
	
	blockChain.Mining(nil)
	blockChain.UpdateBlockchainNetwork()
	_ = time.AfterFunc(time.Second * 15,blockChain.CronMining)
}

func (blockChain *BlockChain) ValidChain(chain []*Block) bool {
	preBlock := chain[0]
	index := 1
	for index<len(chain){
		block := chain[index]
		if block.previousBlockAddress != preBlock.getHash() {
			return false
		}
		//if !blockChain.ValidProof(block.nonce,block.previousBlockAddress,
		//	block.Transactions,MINING_LEVEL){
		//	return false
		//}
		index+=1
		preBlock = block
	}
	return true
}

func (blockChain *BlockChain) UpdateBlockchainNetwork() bool {
	
	var longestChain []*Block = nil
	maxLen := len(blockChain.Chain())
	for _, neighbor := range blockChain.neighbors{
	  	address := fmt.Sprintf("http://%s/chain",neighbor)
	  	log.Println("trying to connect to ",address)
	  	response, err := http.Get(address)
	  	if err==nil && response.StatusCode==200{
	  		var tempChain *BlockChain
	  		decoder := json.NewDecoder(response.Body)
	  		err = decoder.Decode(&tempChain)
	  		if err!=nil{
	  			log.Println("Error Occured while loading chain object",err)
			}

	  		if len(tempChain.Chain()) > maxLen {
				// && blockChain.ValidChain(tempChain.Chain())
	  			maxLen = len(tempChain.Chain())
	  			longestChain = tempChain.Chain()
	  		}
		}else{
			fmt.Printf("Not able to connect to the node %s",neighbor)
		}
	}

	if longestChain !=nil{
		blockChain.chain = longestChain
		log.Println("Blockchain replaced for the node ", blockChain.minerBlockChainAddress)
		return true
	}
	//log.Println("Block chain replace not required")
	return false
}

	//ProofOfWork returns the nonce value for the block
func (blockChain *BlockChain) ProofOfWork(wallet *wallet.Wallet) int {
	   transactions := blockChain.CopyTransactionPool(wallet)
	   previousHash := blockChain.GetLastBlock().getHash()
	   nonce := 0
	   for !blockChain.ValidProof(nonce,previousHash,transactions,MINING_LEVEL){
	   		nonce+=1
	   }
	   return nonce
}

func CreateBlockChain(minerBlockChainAddress string,port uint16) *BlockChain  {
	block := &Block{}
	blockChain := new(BlockChain)
	blockChain.port= port
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

func (blockChain *BlockChain) GetLastBlock() *Block {
	return blockChain.chain[len(blockChain.chain)-1]
}

func (blockchain *BlockChain) VerifyTransaction(
	senderPublicKey *ecdsa.PublicKey,
	signature *utils.Signature,
	transaction *utils.Transaction) bool{
	bytes,_ := json.Marshal(transaction)
	hash := sha256.Sum256(bytes)
	return ecdsa.Verify(senderPublicKey,hash[:],signature.R, signature.S)
}

func (blockChain *BlockChain) CalculateTotalAmount(blockChainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, block := range blockChain.chain  {
		for _, transaction := range block.Transactions {
		  if blockChainAddress==transaction.RecipientBlockchainAddress() {
		  	totalAmount+=transaction.Amount()
		  }
		  if blockChainAddress == transaction.SenderBlockchainAddress() {
		  	totalAmount-=transaction.Amount()
		  }
		}
	}
	return totalAmount
}

func (blockChain *BlockChain) IsTransactionValid(sender string, senderPublicKey *ecdsa.PublicKey,
		signature *utils.Signature, transaction *utils.Transaction) bool {

	// ignoring this for the mining sender, reward is being sent
	if sender == MINING_SENDER_ADDRESS{
		return true
	}
	if !blockChain.VerifyTransaction(senderPublicKey,
		signature,transaction){
		log.Println("Error Invalid Transaction")
		 return false
	}
	
	return true
}

//func main()  {
//	log.SetPrefix("BlockUsingGo\t")
//	minerBlockChainAddress := "BlockChain Miner Address"
//	blockChain := createBlockChain(minerBlockChainAddress)
//	fmt.Println(blockChain.getLastBlock().previousBlockAddress)
//	var transactions []*utils.Transaction
//	nonce := 0
//	transaction := createTransaction("abc","xyz",1.32)
//	transactions, nonce = VerifyAndCreateTransaction(blockChain, transaction, transactions, nonce)
//
//	transactions = nil
//
//	transaction = createTransaction("xyz", "pqr", 5.12)
//	transactions, nonce = VerifyAndCreateTransaction(blockChain, transaction, transactions, nonce)
//
//	blockChain.toString()
//
//	fmt.Println("\nTotal Value for miner address is ",
//		blockChain.calculateTotalAmount(minerBlockChainAddress))
//}

func (blockChain *BlockChain) VerifyAndAddTransaction(wallet *wallet.Wallet ,
	transaction *utils.Transaction,
	signature *utils.Signature) bool {
	

	if blockChain.IsTransactionValid(transaction.SenderBlockchainAddress(),
		wallet.PublicKey(),signature , transaction) {
		
		//if blockChain.CalculateTotalAmount(transaction.SenderBlockchainAddress()) < transaction.Amount(){
		//	log.Printf("Error Insufficient amount in the wallet")
		//	return false
		//}
		var transactions []*utils.Transaction
		transactions = append(transactions, transaction)
		nonce := blockChain.ProofOfWork(wallet)
		blockChain.addBlock(blockChain.GetLastBlock().getHash(), nonce, transactions)
		/* commenting mining as this needs to get automated and
		should not performed after every transaction
		blockChain.Mining(wallet)
		*/

		for _, neighbor := range blockChain.neighbors{
			status := blockChain.sendTransactionToBlockchainNode(wallet, transaction, neighbor)
			if !status {
				//Need to handle gracefully
				break
			}
		}
		return true
	}
	return false
}

func (blockChain *BlockChain) sendTransactionToBlockchainNode(wallet *wallet.Wallet, transaction *utils.Transaction, neighbor string) bool {
	log.Println("Updating the transaction details in ", neighbor)
	privateKey := wallet.PrivateKeyStr()
	publicKey := wallet.PublicKeyStr()
	senderBlockchainAddress := transaction.SenderBlockchainAddress()
	recipientBlockchainAddress := transaction.RecipientBlockchainAddress()
	amount := transaction.Amount()
	request := utils.TransactionInternalRequest{
		SenderPrivateKey:           &privateKey,
		SenderPublicKey:            &publicKey,
		SenderBlockchainAddress:    &senderBlockchainAddress,
		RecipientBlockchainAddress: &recipientBlockchainAddress,
		Amount:                     &amount,
	}
	bytes, _ := json.Marshal(request)
	req_bytes := bytes2.NewBuffer(bytes)
	client := &http.Client{}
	endpoint := fmt.Sprintf(INTERNAL_TRANSACTION_API, neighbor)
	req, _ := http.NewRequest("PUT", endpoint, req_bytes)
	response, _ := client.Do(req)
	log.Println("Transaction Updated status-", response.StatusCode)
	if response.StatusCode == 201 {
		return true
	}else{
		return false
	}
}


func (blockChain *BlockChain) UpdateTransaction(wallet *wallet.Wallet ,
		transaction *utils.Transaction,
		signature *utils.Signature) bool {
	
	if blockChain.IsTransactionValid(transaction.SenderBlockchainAddress(),
		wallet.PublicKey(),signature , transaction) {

		//if blockChain.CalculateTotalAmount(transaction.SenderBlockchainAddress()) < transaction.Amount(){
		//	log.Printf("Error Insufficient amount in the wallet")
		//	return false
		//}
		var transactions []*utils.Transaction
		transactions = append(transactions, transaction)
		nonce := blockChain.ProofOfWork(wallet)
		blockChain.addBlock(blockChain.GetLastBlock().getHash(), nonce, transactions)
		/* commenting mining as this needs to get automated and
		should not performed after every transaction
		blockChain.Mining(wallet)
		*/

		return true
	}
	return false
}

