package server

import (
	"blockchaingo/server"
	"blockchaingo/utils"
	wallet2 "blockchaingo/wallet"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"text/template"
)


const templatePath = "/usr/local/go/src/blockchaingo/wallet/server/templates"
type WalletServer struct {
	port uint16
	gateway	string
}

func init()  {
	log.SetPrefix("Wallet Server\t")
}

func (walletServer *WalletServer) Port() uint16 {
	return walletServer.port
}

func (walletServer *WalletServer) Gateway() string {
	return walletServer.gateway
}

func NewWalletServer(port uint16,gateway string) *WalletServer {
	return &WalletServer{port,gateway}
}

func (walletServer *WalletServer) Wallet(w http.ResponseWriter,r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		  w.Header().Add("Content-Type","application/json")
		  wallet := wallet2.CreateWallet()
		  bytes := wallet.ToByteArray()
		  io.WriteString(w,string(bytes[:]))
	default:
		log.Println("Invalid Http request")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (walletserver *WalletServer) Index(w http.ResponseWriter,r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		templateDir,_ := filepath.Abs(templatePath)
		file,err := template.ParseFiles(path.Join(templateDir,"index.html"))
		if err!=nil{
			log.Println("file not found in the dir \t",templateDir)
		}else{
			file.Execute(w,"")
		}

	default:
		log.Println("Invalid Http request")
	}
}

func (walletserver *WalletServer) Transaction(w http.ResponseWriter,r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var req utils.TransactionRequest
		err := decoder.Decode(&req)
		if err!=nil || !req.Validate(){
			log.Println("Invalid Http request body")
			io.WriteString(w, utils.ToString("failed"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		amount, err := strconv.ParseFloat(*req.Amount,32)
		if err!=nil{
			log.Println("Error while parsing amount")
			io.WriteString(w,utils.ToString("fail"))
			return
		}
		amount32 := float32(amount)

		w.Header().Add("Content-Type","application/json")

		transactionReq := &utils.TransactionInternalRequest{
			SenderBlockchainAddress: req.SenderBlockchainAddress,
			RecipientBlockchainAddress: req.RecipientBlockchainAddress,
			SenderPublicKey: req.SenderPublicKey,
			SenderPrivateKey: req.SenderPrivateKey, Amount:&amount32,
		}

		buffer := bytes.NewBuffer(transactionReq.ToByteArray())
		response, err := http.Post(walletserver.Gateway()+"/internal/transaction",
			"application/json",buffer)
		
		log.Println("Internal API response ",response.StatusCode)
		if response.StatusCode ==201 {
			io.WriteString(w,utils.ToString("transaction success"))
			return
		}
		io.WriteString(w,utils.ToString("fail"))

	default:
		log.Println("Invalid Http request")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (walletServer *WalletServer) Run(blockChainServer *server.BlockchainServer) {
	log.Println("WalletServer and blockchain server Started")
	blockChainServer.GetOrCreateChain().Run()
	// Wallet APIs
	http.HandleFunc("/",walletServer.Index)
	http.HandleFunc("/wallet",walletServer.Wallet)
	http.HandleFunc("/wallet/amount",blockChainServer.GetAmount)
	http.HandleFunc("/transaction",walletServer.Transaction)
	// Blockchain server APIs
	http.HandleFunc("/chain",blockChainServer.GetChain)
	http.HandleFunc("/transactions",blockChainServer.GetTransactions)
	http.HandleFunc("/internal/transaction",blockChainServer.AddTransactions)
	http.HandleFunc("/mine",blockChainServer.Mining)
	http.HandleFunc("/mine/cron",blockChainServer.CronMining)
	
	server_address := "localhost:"+strconv.Itoa(int(walletServer.Port()))
	log.Fatal(http.ListenAndServe(server_address,nil))
}

