package server

import (
	"flag"
	"log"
)

func init()  {
	log.SetPrefix("BlockchainServer:")
}

func main()  {
	port := flag.Uint("port",5000,"Server TCP port")
	server := NewBlockchainServer(uint16(*port))
	server.Run()
}
