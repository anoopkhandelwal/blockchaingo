package main

import (
	"blockchaingo/server"
	ws "blockchaingo/wallet/server"
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Main\t")
}

func mainBlockchainServer()  {
	port := flag.Uint("port",5001,"BlockchainServer TCP port")
	blockchainServer := server.NewBlockchainServer(uint16(*port))
	blockchainServer.Run()
}

func mainWalletServer()  {
	port := flag.Uint("port",5000,"WalletServer TCP port")
	flag.Parse()
	gateway := flag.String("gateway","http://127.0.0.1:5000","Gateway")
	walletServer := ws.NewWalletServer(uint16(*port),*gateway)
	blockchainServer := server.NewBlockchainServer(uint16(*port))
	walletServer.Run(blockchainServer)
}

func main()  {
	 mainWalletServer()
}

