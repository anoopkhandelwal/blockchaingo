package utils

import (
	"encoding/json"
	"log"
)

type TransactionRequest struct {
	SenderPrivateKey *string	`json:"sender_private_key"`
	SenderBlockchainAddress	*string	`json:"sender_blockchain_address"`
	RecipientBlockchainAddress	*string	`json:"recipient_blockchain_address"`
	SenderPublicKey *string	`json:"sender_public_key"`
	Amount *string	`json:"amount"`
}

func (request *TransactionRequest) Validate() bool {
	if request.SenderPrivateKey == nil ||
		request.SenderBlockchainAddress == nil ||
		request.Amount == nil ||
		request.SenderPublicKey == nil ||
		request.RecipientBlockchainAddress == nil{
		return false
	}
	return true
}

type TransactionInternalRequest struct {
	SenderPrivateKey *string	`json:"sender_private_key"`
	SenderBlockchainAddress	*string	`json:"sender_blockchain_address"`
	RecipientBlockchainAddress	*string	`json:"recipient_blockchain_address"`
	SenderPublicKey *string	`json:"sender_public_key"`
	Amount *float32	`json:"amount"`
}

func (request *TransactionInternalRequest) Validate() bool {
	if request.SenderBlockchainAddress == nil ||
			request.Amount == nil ||
			request.SenderPublicKey == nil ||
			request.RecipientBlockchainAddress == nil{
		return false
	}
	return true
}

func (request *TransactionInternalRequest) ToByteArray() []byte {
	bytes,err := json.Marshal(request)
	if err!=nil{
		log.Println("Error Occured in marshalling")
		return nil
	}
	return bytes
}
