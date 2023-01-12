package main

import (
	"blockchaingo/utils"
	"log"
)

func init()  {
	log.SetPrefix("Cmd Main\t")
}


func main() {
	log.Println("Hosts Discovery ",utils.FindBlockchainNeighbors("127.0.0.1",5000, 0,3,5000,5003))

	log.Println("HostsName ",utils.GetHost())
	
}
