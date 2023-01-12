package utils

import (
	"fmt"
	"log"
	"math/big"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

func init()  {
	log.SetPrefix("Signature\t")
}
func (signature *Signature) ToStr() string {
	return fmt.Sprintf("Signature\t%064x%064x",signature.R.Bytes(), signature.S.Bytes())
}

func DecodeSignature(signature string) *Signature {
	r,s := DecodeStr(signature)
	return &Signature{R:&r,S:&s}
}


