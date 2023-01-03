package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
)

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	publicKey *ecdsa.PublicKey
}

func (wallet *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return wallet.privateKey
}

func (wallet *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x",wallet.privateKey.D.Bytes())
}

func (wallet *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%x%x",wallet.publicKey.X.Bytes(),wallet.publicKey.Y.Bytes())
}

func (wallet *Wallet) PublicKey() *ecdsa.PublicKey {
	return wallet.publicKey
}


func (wallet *Wallet) SetPublicKey(publicKey *ecdsa.PublicKey) {
	wallet.publicKey = publicKey
}

func CreateWallet() *Wallet {
	  wallet := new(Wallet)
	  privateKey, _ := ecdsa.GenerateKey(elliptic.P256(),rand.Reader)
	  wallet.privateKey = privateKey
	  wallet.publicKey = &wallet.privateKey.PublicKey
	  return wallet
}

func init()  {
	log.SetPrefix("BlockUsingGoWallet")
}

func main() {
	fmt.Println("Staritng creating wallet")
	wal := CreateWallet()
	fmt.Println(wal.PrivateKeyStr())
	fmt.Println(wal.PublicKeyStr())
}

