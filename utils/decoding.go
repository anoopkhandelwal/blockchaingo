package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"
)
func DecodeStr(signature string) (big.Int,big.Int) {

	bytesX,_ := hex.DecodeString(signature[:64])
	bytesY,_ := hex.DecodeString(signature[64:])
	var x big.Int
	var y big.Int
	_ = x.SetBytes(bytesX)
	_ = y.SetBytes(bytesY)
	return x,y
}

func DecodePublicKey(publicKey string) *ecdsa.PublicKey {
	x,y := DecodeStr(publicKey)
	return &ecdsa.PublicKey{elliptic.P256(),&x,&y}
}

func DecodePrivateKey(privateKey string, publicKey *ecdsa.PublicKey) *ecdsa.PrivateKey {

	bytes ,_ :=hex.DecodeString(privateKey[:])
	var key big.Int
	_ = key.SetBytes(bytes)
	return &ecdsa.PrivateKey{*publicKey,&key}
}



