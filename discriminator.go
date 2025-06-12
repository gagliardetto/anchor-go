package main

import (
	"crypto/sha256"
	"github.com/gagliardetto/utilz"
)

func Discriminator(preimage string) *[8]byte {
	hash := sha256.Sum256([]byte(preimage))
	var result [8]byte
	copy(result[:], hash[:8])
	return &result
}

func AccountDiscriminator(name string) *[8]byte {
	return Discriminator("account:" + utilz.ToCamel(name))
}

func EventDiscriminator(name string) *[8]byte {
	return Discriminator("event:" + name)
}
