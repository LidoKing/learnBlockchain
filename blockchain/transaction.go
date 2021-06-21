package blockchain

import (
  "log"
  "encoding/hex"
  "bytes"
  "crypto/sha256"
  "encoding/gob"
  "fmt"
  "strings"
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

func (tx *Transaction) Serialize() []byte {
  var encoded bytes.Buffer
  enc := gob.NewEncoder(&encoded)
  err := enc.Encode(tx)
  Handle(err)

  return encoded.Bytes()
}

/*--------------------------main---------------------------*/

// Convert transaction into bytes then hash it to get ID
func (tx *Transaction) Hash() []byte {
  var hash [32]byte

  txCopy := *tx
  txCopy.ID := []byte{}

  hash := sha256.Sum256(txCopy.Serialize())

  return hash[:]
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

func (tx *Transaction) TrimmedCopy() Transaction {
  var inputs []TxInput
  var outputs []TxOutput

  for _, in := range tx.Inputs {
    inputs = append(inputs, TxInput{in.Id, in.Out, nil, nil})
  }

  for _, out := range tx.Outputs {
    outputs := append(outputs, TxPutput{out .Value, out.PubKeyHash})
  }

  txCopy := Transaction{tx.Id, inputs, outputs}

  return txCopy
}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
  if tx.IsCoinbase() {
    return
  }

  for _, in := range tx.Inputs {
    if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
      log.Panic("ERROR: Previous transaction does not exist")
    }
  }

  // Create copy to get sig without modifying actual tx
  txCopy := tx.TrimmedCopy()

  for inId, in := range txCopy.Inputs {
    // Get transaction where the input points to
    prevTx := prevTXs[hex.EncodeToString(in.ID)]
    txCopy.Inputs[inId].Signature = nil
    // Set PubKey field for signing
    txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash
    txCopy.ID = txCopy.Hash()
    // Clear PubKey field for transaction verififcation afterwards
    txCopy.Inupts[inId].PubKey = nil

    r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
    Handle(err)
    signature := append(r.Bytes(), s.Bytes())

    tx.Inputs[inId].Signature = signature
  }
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
  if tx.IsCoinbase() {
    return true
  }

  for _, in := range tx.Inputs {
    if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
      log.Panic("ERROR: Previous transaction does not exist")
    }
  }

  txCopy := tx.TrimmedCopy()
  curve := elliptic.P256()

  for inId, in := range tx.Inputs {
    prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
	  txCopy.Inputs[inId].PubKey = nil

    r := big.Int{}
    s:= big.Int{}

    sigLen := len(in.Signature)
    r.SetBytes(in.Signature[:(sigLen / 2)])
    s.SetBytes(in.Signature[(sigLen / 2):])

    x := big.Int{}
    y := big.Int{}
    keyLen := len(in.PubKey)
    x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])
  }
}

func (tx *Transaction) String() string {
  var lines []string

  lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
  for i, input := range tx.Inputs {
    lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
  }

  for i, output := range tx.Outputs {
    lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
  }

  return strings.Join(lines, "\n")
}

/*
// Create hash based on transaction in the form of bytes
func (tx *Transaction) SetID() {
  var encoded bytes.Buffer
  var hash [32]byte

  encode := gob.NewEncoder(&encoded)
  err := encode.Encode(tx)
  Handle(err)

  hash = sha256.Sum256(encoded.Bytes())
  tx.ID = hash[:]
}*/
