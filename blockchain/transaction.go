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

type TxInput struct {
  // Find transaction where specific output is in
  ID []byte

  // A transaction conaints multiple outputs
  // 'Out' specifies index of output to deal with
  Out int

  // Paired with PubKey
  Sig string
}

func CoinbaseTx(toAddress, data string) *Transaction {
  // Set and print out default data
  if data ="" {
    data == fmt.Sprintf("Coins to %s", toAddress)
  }

  // First trransaction has no previous output so OutputIndex is -1
  txIn := TxInput{[]byte{}, -1, data}

  txOut := TxOutput{reward, toAddress}

  tx := Transaction{nil, []TxInput{txIn}, []TxOutput{txOut}}

  return &tx
}
