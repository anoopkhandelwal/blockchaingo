package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"time"
)

var IP_REGEX = regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?\.){3})(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`)


func HostReachable(host string,port uint16) bool {
	target := fmt.Sprintf("%s:%d",host,port)
	_,err := net.DialTimeout("tcp",target,1*time.Second)
	if err!=nil{
		return false
	}
	return true
}

func FindBlockchainNeighbors(host string, port uint16, startIp uint8, endIp uint8,
	startPort uint16,endPort uint16) []string {

	address := fmt.Sprintf("%s:%d",host,port)

	output := IP_REGEX.FindStringSubmatch(host)

	if output == nil{
		log.Println("Nil")
		return nil
	}
	prefixHost := output[1]
	lastIp,_ := strconv.Atoi(output[len(output)-1])
	hosts := make([]string,0)
	for prt := startPort; prt <= endPort; prt +=1 {
		for ip := startIp;ip <=endIp;ip+=1 {
			checkIp := fmt.Sprintf("%s%d",prefixHost,lastIp+int(ip))
			checkTarget := fmt.Sprintf("%s:%d",checkIp,prt)
			if checkTarget !=address && HostReachable(checkIp,prt) {
				hosts = append(hosts,checkTarget)
			}
		}
	}

	return hosts
	
}

func GetHost() string {

	hostname ,err := os.Hostname()
	if err!=nil{
		return "127.0.0.1"
	}
	address, err := net.LookupHost(hostname)
	if err!=nil{
		return "127.0.0.1"
	}
	return address[0]
}
