package wallet

import (
  "crypto/ecdsa"
  "crypto/elliptic"
  "crypto/rand"
  "crypto/sha256"
  "log"

  "golang.org/x/crypto/ripemd160"
  "github.com/LidoKing/learnBlockchain/blockchain"
)

const (
  checksumLength = 4
  version = byte{0x00}
)

type Wallet struct {
  PrivateKey ecdsa.PrivateKey
  PublicKey  []byte
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
  curve := elliptic.P256()

  private, err := ecdsa.GenerateKey(curve, rand.Reader)
  Handle(err)

  // concatenate two slices of bytes, append(x, y...)
  public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

  return *private, public
}

func MakeWallet() *Wallet {
  private, public := NewKeyPair()
  wallet := Wallet{private, public}

  return &wallet
}
