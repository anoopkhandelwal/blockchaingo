package server

import (
	"blockchaingo/blockchain"
	"blockchaingo/utils"
	"blockchaingo/wallet"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)
var cache map[string]*blockchain.BlockChain = make(map[string]*blockchain.BlockChain)

func init()  {
	log.SetPrefix("Blockchain Server\t")
}

type BlockchainServer struct {
	port uint16
}

func (b *BlockchainServer) Port() uint16 {
	return b.port
}

func NewBlockchainServer(port uint16) *BlockchainServer  {
	return &BlockchainServer{port}
}


func (blockchainServer *BlockchainServer) GetOrCreateChain() *blockchain.BlockChain {

	_, ok := cache["blockchain"]
	if !ok{
		log.Println("Creating the blockchain")
		walletMiner := wallet.CreateWallet()
		bc := blockchain.CreateBlockChain(walletMiner.BlockchainAddress(),blockchainServer.Port())
		cache["blockchain"] = bc
	}

	return cache["blockchain"]
}

func (server *BlockchainServer) GetChain(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		blockchain := server.GetOrCreateChain()
		bytes, _ := blockchain.MarshalJSON()
		log.Println("Valid chain ",blockchain.ValidChain(blockchain.Chain()))
		w.Header().Add("Content-Type","application/json")
		io.WriteString(w, string(bytes[:]))
	default:
		log.Println("Invalid Http request")
		
	}
}

func (blockchainServer *BlockchainServer) AddTransactions(w http.ResponseWriter, r *http.Request)  {
	switch r.Method {

	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var req utils.TransactionInternalRequest
		err := decoder.Decode(&req)
		if err != nil || !req.Validate() {
			log.Printf("Error %v",err)
			io.WriteString(w,utils.ToString("fail"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		publicKey := utils.DecodePublicKey(*req.SenderPublicKey)
		privateKey := utils.DecodePrivateKey(*req.SenderPrivateKey,publicKey)
		blockChain := blockchainServer.GetOrCreateChain()

		transaction := utils.CreateTransaction(privateKey,
			publicKey,*req.SenderBlockchainAddress,*req.RecipientBlockchainAddress,
			*req.Amount)
		signature := transaction.GenerateSignature()
		sender := wallet.CreateWalletWithKeys(privateKey,publicKey)
		isAdded := blockChain.VerifyAndAddTransaction(sender,transaction,signature)
		log.Println("isAdded? ",isAdded)
		w = utils.AddContentType(w)
		if !isAdded{
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w,utils.ToString("fail"))
		}else{
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w,utils.ToString("success"))
		}

	case http.MethodPut:
		log.Println("Update Transaction Request received")
		decoder := json.NewDecoder(r.Body)
		var req utils.TransactionInternalRequest
		err := decoder.Decode(&req)
		if err != nil || !req.Validate() {
			log.Printf("Error %v",err)
			io.WriteString(w,utils.ToString("fail"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		publicKey := utils.DecodePublicKey(*req.SenderPublicKey)
		privateKey := utils.DecodePrivateKey(*req.SenderPrivateKey,publicKey)
		blockChain := blockchainServer.GetOrCreateChain()

		transaction := utils.CreateTransaction(privateKey,
			publicKey,*req.SenderBlockchainAddress,*req.RecipientBlockchainAddress,
			*req.Amount)
		signature := transaction.GenerateSignature()
		sender := wallet.CreateWalletWithKeys(privateKey,publicKey)
		isUpdated := blockChain.UpdateTransaction(sender,transaction,signature)
		log.Println("isUpdated? ",isUpdated)
		w = utils.AddContentType(w)
		if !isUpdated{
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w,utils.ToString("fail"))
		}else{
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w,utils.ToString("success"))
		}
	case http.MethodDelete:
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w,utils.ToString("success"))

	default:
		log.Println("Invalid Http request")
		w.WriteHeader(http.StatusBadRequest)
	}
}

type TransactionList struct {
	Transactions []*utils.Transaction	`json:"transactions"`
}


func (blockchainServer *BlockchainServer) GetTransactions(w http.ResponseWriter, r *http.Request)  {
	switch r.Method {
	case http.MethodGet:
		w = utils.AddContentType(w)
		blockChain := blockchainServer.GetOrCreateChain()
		transactions := make([]*utils.Transaction ,0)

		for _,block := range blockChain.Chain(){
			ts := block.Transactions
			for _,t := range ts{
				transactions = append(transactions,t)
			}
		}

		bytes, _ := json.Marshal(struct {
			Transactions 	[]*utils.Transaction	`json:"transactions"`
			Size			int						`json:"size"`

		}{
			Transactions: transactions,
			Size: len(transactions),
		})

		io.WriteString(w,string(bytes[:]))

	default:
		log.Println("Invalid Http request")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (blockchainServer *BlockchainServer) Mining(w http.ResponseWriter, r *http.Request){
	switch r.Method {
	case http.MethodGet:
		  blockchain := blockchainServer.GetOrCreateChain()

		  // Sample wallet created.Ideally it wont be a required attribute
		  wallet := wallet.CreateWallet()
		  isMined := blockchain.Mining(wallet)
		  w = utils.AddContentType(w)
		  if !isMined {
		  	log.Println("Mining failed")
		  	w.WriteHeader(http.StatusInternalServerError)
		  	io.WriteString(w,utils.ToString("failed"))
		  } else{
			  w.WriteHeader(http.StatusAccepted)
			  io.WriteString(w,utils.ToString("success"))
		  }
	
	default:
		log.Println("Invalid Http request")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (blockchainServer *BlockchainServer) CronMining(w http.ResponseWriter, r *http.Request){
	switch r.Method {
	case http.MethodGet:
		blockchain := blockchainServer.GetOrCreateChain()
		if len(blockchain.Chain())>1{blockchain.CronMining()
			w.WriteHeader(http.StatusAccepted)
			io.WriteString(w,utils.ToString("success"))
		} else{
			io.WriteString(w,utils.ToString("failed"))
	}

	default:
		log.Println("Invalid Http request")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (blockchainServer *BlockchainServer) GetAmount(w http.ResponseWriter, r *http.Request)  {
	switch r.Method {
	case http.MethodGet:
		w = utils.AddContentType(w)
		blockChainAddress := r.URL.Query().Get("blockchain_address")
		amount := blockchainServer.GetOrCreateChain().CalculateTotalAmount(blockChainAddress)
		
		bytes, _ := json.Marshal(struct {
			Amount float32 `json:"amount"`
		}{
			Amount: amount,
		})
		io.WriteString(w,string(bytes[:]))
		
	default:
		log.Println("Invalid Http request")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (b *BlockchainServer) Run() {
	log.Println("Blockchain server Started")
	b.GetOrCreateChain().Run()
	http.HandleFunc("/",b.GetChain)
	server_address := "localhost:"+strconv.Itoa(int(b.Port()))
	log.Fatal(http.ListenAndServe(server_address,nil))
}

