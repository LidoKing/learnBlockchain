 package blockchain

import (
  "bytes"
  "encoding/gob"
  "github.com/LidoKing/learnBlockchain/blockchain/wallet"
)

type TxOutput struct {
  // Representative of the amount of tokens in a transaction
  Value int

  // Hash for the address taht 'owns' the output
  PubKeyHash []byte
}

type TxOutputs struct {
  Outputs []TxOutput
}

// Reference to previous TxOutput
// Transactions that have outputs, but no inputs pointing to them are spendable (UTXOs)
type TxInput struct {
  // Points to transaction where output is in
  ID []byte
  // Index of output that input points to
  Out int

  Sig []byte
  PubKey []byte
}

func (outs TxOutputs) Serialize() []byte {
  var buffer bytes.Buffer
  enc := gob.NewEncoder(&buffer)
  err := enc.Encode(outs)
  Handle(err)
  return buffer.Bytes()
}

func DeserializeOutputs(data []byte) TxOutputs {
  var outputs TxOutputs
  dec := gob.NewDecoder(bytes.NewReader(data))
  err := dec.Decode(&outputs)
  Handle(err)
  return outputs
}

// Check if unspent outputs belong to specific address
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
  return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// Lock output with address
func (out *TxOutput) Lock(address []byte) {
  pubKeyHash := wallet.Base58Decode(address)

  // Remove version and checksum
  pubKeyHash = pubKeyHash[1:len(pubKeyHash)-4]
  out.PubKeyHash = pubKeyHash
}

func NewTXOutput(value int, address string) *TxOutput {
  txo := &TxOutput{value, nil}
  txo.Lock([]byte(address))

  return txo
}
