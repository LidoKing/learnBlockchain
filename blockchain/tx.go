package blockchain

import (
  "github.com/LidoKing/learnBlockchain/blockchain/wallet"
)

type TxOutput struct {
  // Representative of the amount of tokens in a transaction
  Value int

  // Unlock tokens for transaction
  PubKeyHash []byte
}

// Reference to previous TxOutput
// Transactions that have outputs, but no inputs pointing to them are spendable (UTXOs)
type TxInput struct {
  // Point to transaction where specific output is in
  ID []byte

  // A transaction has multiple outputs
  // 'Out' specifies index of output to deal with
  Out int
  Sig []byte
  PubKey []byte
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
  lockingHash := wallet.PublicKeyHash(in.PubKey)

  return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// Lock output with address
func (out *TxOutput) Lock(address []byte) {
  pubKeyHash := wallet.Base58Decode(address)

  // Remove version and checksum
  pubKeyHash = pubKeyHash[1:len(pubKeyHash)-4]
  out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
  return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(value int, address string) *TxOutput {
  txo := &TxOuput{value, nil}
  txo.Lock([]byte(address))

  return txo
}
