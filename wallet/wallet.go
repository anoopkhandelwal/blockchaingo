package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/ripemd160"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"log"
)

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	publicKey *ecdsa.PublicKey
	blockchainAddress string
}

func (wallet *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return wallet.privateKey
}

func (wallet *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x",wallet.privateKey.D.Bytes())
}

func (wallet *Wallet) ToByteArray() []byte {
	bytes, _ := json.Marshal(struct {
		PrivateKey	string	`json:"private_key"`
		PublicKey	string	`json:"public_key"`
		BlockchainAddress	string	`json:"blockchain_address"`
	}{
		PrivateKey:wallet.PrivateKeyStr(),
		PublicKey: wallet.PublicKeyStr(),
		BlockchainAddress: wallet.BlockchainAddress(),
	})
	if bytes!=nil{
		return bytes
	}else{
		log.Println("Not able to marshal")
		return nil
	}
}

func (wallet *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%064x%064x",wallet.publicKey.X.Bytes(),wallet.publicKey.Y.Bytes())
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
	  wallet.CreateBlockchainAddress()
	  return wallet
}

func CreateWalletWithKeys(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) *Wallet {

	wallet := new(Wallet)
	wallet.privateKey = privateKey
	wallet.publicKey = publicKey
	wallet.CreateBlockchainAddress()
	return wallet

}

func (wallet *Wallet) CreateBlockchainAddress() {
	hash := sha256.New()
	hash.Write(wallet.publicKey.X.Bytes())
	hash.Write(wallet.publicKey.Y.Bytes())
	digest := hash.Sum(nil)

	hash2 := ripemd160.New()
	hash2.Write(digest)
	digest2 := hash2.Sum(nil)

	versionDigest := make([]byte,21)
	versionDigest[0] = 0x00
	copy(versionDigest[1:],digest2[:])

	hash3 := sha256.New()
	hash3.Write(versionDigest)
	versionDigest2 := hash3.Sum(nil)

	hash4 := sha256.New()
	hash4.Write(versionDigest2)
	versionDigest3 := hash4.Sum(nil)

	checksum := versionDigest3[:4]
	checkSumDigest := make([]byte,25)
	copy(checkSumDigest[:21],versionDigest[:])
	copy(checkSumDigest[21:],checksum)
	wallet.blockchainAddress = base58.Encode(checkSumDigest)
}

func (wallet *Wallet) SetBlockchainAddress(address string)  {
	wallet.blockchainAddress = address
}

func (wallet *Wallet) BlockchainAddress() string {
	return wallet.blockchainAddress
}

func init(){
	log.SetPrefix("Wallet Class\t")
}
