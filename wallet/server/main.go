package server

import (
	"flag"
	"log"
)

func init()  {
	log.SetPrefix("Wallet Server Main-2\t")
}

func main()  {
	port := flag.Uint("port",5000,"Server TCP port")
	gateway := flag.String("gateway","http://127.0.0.1:5000","Gateway")
	server := NewWalletServer(uint16(*port),*gateway)
	server.Run(nil)
}

