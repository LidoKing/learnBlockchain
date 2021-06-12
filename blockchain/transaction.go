package blockchain

import (
  "log"
  "encoding/hex"
  "bytes"
  "crypto/sha256"
  "encoding/gob"
  "fmt"
)

const reward = 100

type Transaction struct {
  ID      []byte
  Inputs  []TxInput
  Outputs []TxOutput
}

/*--------------------------utils---------------------------*/

func (tx *Transaction) IsCoinbase() bool {
  return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

/*--------------------------main---------------------------*/

// Create hash based on transaction in the form of bytes
func (tx *Transaction) SetID() {
  var encoded bytes.Buffer
  var hash [32]byte

  encode := gob.NewEncoder(&encoded)
  err := encode.Encode(tx)
  Handle(err)

  hash = sha256.Sum256(encoded.Bytes())
  tx.ID = hash[:]
}

func CoinbaseTx(toAddress string, data string) *Transaction {
  // Set and print out default data
  if data == "" {
    data = fmt.Sprintf("Coins to %s", toAddress)
  }

  // First trransaction has no previous output
  // OutputIndex is -1
  txIn := TxInput{[]byte{}, -1, data}
  txOut := TxOutput{reward, toAddress}

  tx := Transaction{nil, []TxInput{txIn}, []TxOutput{txOut}}
  tx.SetID()

  return &tx
}

func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
  var inputs []TxInput
  var outputs []TxOutput

  // vallidOutputs is a map!!! (with stringified transaction IDs as keys)
  spendable, validOutputs := chain.FindSpendableOutputs(from, amount)

  if spendable < amount {
    log.Panic("Error: Not enough funds!")
  }

  // txid is key of map (string), outs is index of output that is unspent
  for txid, outs := range validOutputs {
    // Convert stringified IDs back to slice of bytes
    txID, err := hex.DecodeString(txid)
    Handle(err)

    // Create inputs of new transaction that points to to-be-used UTXOs
    for _, out := range outs {
      input := TxInput{txID, out, from}
      inputs = append(inputs, input)
    }
  }

  outputs = append(outputs, TxOutput{amount, to})

  // Send change back to sender, i.e. new UTXO
  if spendable > amount {
    outputs = append(outputs, TxOutput{spendable - amount, from})
  }

  tx := Transaction{nil, inputs, outputs}
  tx.SetID()

  return &tx
}
