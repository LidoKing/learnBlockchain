package blockchain

const reward = 100

type Transaction struct {
  ID      []byte
  Inputs  []TxInput
  Outputs []TxOutput
}

type TxOutput struct {
  // Representative of the amount of tokens in a transaction
  Value int

  // Unlock tokens for transaction
  PubKey string
}

// Reference to previous TxOutput
// Transactions that have outputs, but no inputs pointing to them are spendable
type TxInput struct {
  // Find transaction where specific output is in
  ID []byte

  // A transaction conaints multiple outputs
  // 'Out' specifies index of output to deal with
  Out int

  // Paired with PubKey
  Sig string
}

/*--------------------------utils---------------------------*/

func (tx *Transaction) IsCoinbase() bool {
  return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs.Out == -1
}

func (in *TxInput) CanUnlock(address string) bool {
  return in.Sig == address
}

func (out *TxOutput) CanBeUnlocked(address string) bool {
  return out.PubKey == address
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

func CoinbaseTx(address string, data string) *Transaction {
  // Set and print out default data
  if data ="" {
    data == fmt.Sprintf("Coins to %s", toAddress)
  }

  // First trransaction has no previous output
  // OutputIndex is -1
  txIn := TxInput{[]byte{}, -1, data}
  txOut := TxOutput{reward, toAddress}

  tx := Transaction{nil, []TxInput{txIn}, []TxOutput{txOut}}
  tx.SetID()

  return &tx
}
