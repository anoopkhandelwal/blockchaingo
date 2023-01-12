package utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"log"
)

type Transaction struct {
	senderPrivateKey *ecdsa.PrivateKey	`json:"sender_private_key,omitempty"`
	senderPublicKey *ecdsa.PublicKey	`json:"sender_public_key,omitempty"`
	senderBlockchainAddress string	`json:"sender_blockchain_address,omitempty"`
	recipientBlockchainAddress string	`json:"recipient_blockchain_address"`
	amount float32	`json:"amount"`
}

func (transaction *Transaction) SenderPrivateKey() *ecdsa.PrivateKey {
	return transaction.senderPrivateKey
}

func (transaction *Transaction) SenderPublicKey() *ecdsa.PublicKey {
	return transaction.senderPublicKey
}

func (transaction *Transaction) SenderBlockchainAddress() string {
	return transaction.senderBlockchainAddress
}

func (transaction *Transaction) RecipientBlockchainAddress() string {
	return transaction.recipientBlockchainAddress
}

func (transaction *Transaction) Amount() float32 {
	return transaction.amount
}

func (transaction *Transaction) GenerateSignature() *Signature  {

	bytes,_ := json.Marshal(transaction)
	hash := sha256.Sum256(bytes)
	r,s,_ := ecdsa.Sign(rand.Reader,transaction.senderPrivateKey,hash[:])
	return &Signature{R:r,S:s}
}

func (transaction *Transaction) MarshalJSON() ([]byte,error) {
	return json.Marshal(struct {
		SenderPublicKey				*ecdsa.PublicKey	`json:"sender_public_key"`
		SenderPrivateKey			*ecdsa.PrivateKey	`json:"sender_private_key"`
		SenderBlockchainAddress    string  `json:"sender_blockchain_address"`
		RecipientBlockchainAddress string  `json:"recipient_blockchain_address"`
		Amount                     float32 `json:"amount"`

	}{
		SenderBlockchainAddress:    transaction.senderBlockchainAddress,
		RecipientBlockchainAddress: transaction.recipientBlockchainAddress,
		Amount:                     transaction.amount,
		SenderPrivateKey: 			transaction.senderPrivateKey,
		SenderPublicKey: 			transaction.senderPublicKey,
	})
}

func CreateTransaction(privateKey *ecdsa.PrivateKey,
		publicKey *ecdsa.PublicKey, sender string, recipient string, amount float32) *Transaction {
	
	return &Transaction{
		senderPrivateKey:privateKey,
		senderPublicKey: publicKey,
		senderBlockchainAddress: sender,
		recipientBlockchainAddress: recipient,
		amount:amount}
}

func (transaction *Transaction) ToString() {
	log.Println("Sender	\n",transaction.SenderBlockchainAddress())
	log.Println("Receiver	\n",transaction.RecipientBlockchainAddress())
	log.Println("Amount	\n",transaction.Amount())
}

func (transaction *Transaction) UnmarshalJSON(bytes []byte) error {

	strct := &struct {
		SenderBlockchainAddress	*string	`json:"sender_blockchain_address"`
		RecipientBlockchainAddress	*string	`json:"recipient_blockchain_address"`
		SenderPrivateKey	**ecdsa.PrivateKey	`json:"sender_private_key"`
		SenderPublicKey		**ecdsa.PublicKey	`json:"sender_public_key"`
		Amount				*float32	`json:"amount"`
	}{
		SenderBlockchainAddress: &transaction.senderBlockchainAddress,
		RecipientBlockchainAddress: &transaction.recipientBlockchainAddress,
		SenderPublicKey: &transaction.senderPublicKey,
		SenderPrivateKey: &transaction.senderPrivateKey,
		Amount: &transaction.amount,
	}
	if err := json.Unmarshal(bytes, &strct); err!=nil{
		return err
	}
	return nil
}
