package blockchain

import (
  "log"
  "encoding/hex"
  "bytes"
  "crypto/sha256"
  "crypto/ecdsa"
  "crypto/elliptic"
  "crypto/rand"
  "encoding/gob"
  "fmt"
  "strings"
  "math/big"
  "github.com/LidoKing/learnBlockchain/blockchain/wallet"
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

// Convert transaction into bytes then hash it to get ID
func (tx *Transaction) Hash() []byte {
  var hash [32]byte

  txCopy := *tx
  txCopy.ID = []byte{}

  hash = sha256.Sum256(txCopy.Serialize())

  return hash[:]
}

/*--------------------------main---------------------------*/

func CoinbaseTx(toAddress string, data string) *Transaction {
  // Set and print out default data
  if data == "" {
    // Create slice of byte which has a length of 24
    randData := make([]byte, 24)
    // Gen 24 random bytes
    _, err := rand.Read(randData)
    if err != nil {
      log.Panic(err)
    }
    data = fmt.Sprintf("%x", randData)
  }

  // First trransaction has no previous output
  // OutputIndex is -1
  txIn := TxInput{[]byte{}, -1, nil, []byte(data)}
  txOut := NewTXOutput(20, toAddress)

  tx := Transaction{nil, []TxInput{txIn}, []TxOutput{*txOut}}
  tx.ID = tx.Hash()

  return &tx
}

func NewTransaction(from, to string, amount int, UTXO *UTXOSet) *Transaction {
  var inputs []TxInput
  var outputs []TxOutput

  wallets, err := wallet.CreateWallets()
  Handle(err)
  w:= wallets.GetWallet(from)
  pubKeyHash := wallet.PublicKeyHash(w.PublicKey)

  // vallidOutputs is a map!!! (with stringified transaction IDs as keys)
  spendable, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

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
      input := TxInput{txID, out, nil, w.PublicKey}
      inputs = append(inputs, input)
    }
  }

  outputs = append(outputs, *NewTXOutput(amount, to))

  // Send change back to sender, i.e. new UTXO
  if spendable > amount {
    outputs = append(outputs, TxOutput{spendable - amount, pubKeyHash})
  }

  tx := Transaction{nil, inputs, outputs}
  tx.ID = tx.Hash()
  UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)

  return &tx
}

func (tx *Transaction) TrimmedCopy() Transaction {
  var inputs []TxInput
  var outputs []TxOutput

  for _, in := range tx.Inputs {
    inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
  }

  for _, out := range tx.Outputs {
    outputs = append(outputs, TxOutput{out .Value, out.PubKeyHash})
  }

  txCopy := Transaction{tx.ID, inputs, outputs}

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

  // Create copy to get sig without affecting actual tx
  txCopy := tx.TrimmedCopy()

  for inId, in := range txCopy.Inputs {
    // Get transaction that the input points to
    prevTX := prevTXs[hex.EncodeToString(in.ID)]
    txCopy.Inputs[inId].Sig = nil
    // Set PubKey field for hashing
    // Actual tx is hashed with pubKey instead of pubKeyHash and with no signature
    txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash
    txCopy.ID = txCopy.Hash()
    // Clear PubKey field to prevent unecessary errors
    txCopy.Inputs[inId].PubKey = nil

    r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
    Handle(err)
    signature := append(r.Bytes(), s.Bytes()...)

    tx.Inputs[inId].Sig = signature
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
		txCopy.Inputs[inId].Sig = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
	  txCopy.Inputs[inId].PubKey = nil

    r := big.Int{}
    s:= big.Int{}

    sigLen := len(in.Sig)
    r.SetBytes(in.Sig[:(sigLen / 2)])
    s.SetBytes(in.Sig[(sigLen / 2):])

    x := big.Int{}
    y := big.Int{}
    keyLen := len(in.PubKey)
    x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

    rawPubKey := ecdsa.PublicKey{curve, &x, &y}
    if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
      return false
    }
  }
  return true
}

func (tx *Transaction) String() string {
  var lines []string

  lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
  for i, input := range tx.Inputs {
    lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Sig))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
  }

  fmt.Println()

  for i, output := range tx.Outputs {
    lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       PubKeyHash: %x", output.PubKeyHash))
  }

  return strings.Join(lines, "\n")
}
