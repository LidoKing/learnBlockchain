package blockchain

type TxOutput struct {
  // Representative of the amount of tokens in a transaction
  Value int

  // Unlock tokens for transaction
  PubKey string
}

// Reference to previous TxOutput
// Transactions that have outputs, but no inputs pointing to them are spendable (UTXOs)
type TxInput struct {
  // Point to transaction where specific output is in
  ID []byte

  // A transaction conaints multiple outputs
  // 'Out' specifies index of output to deal with
  Out int

  // Paired with PubKey
  Sig string
}

func (in *TxInput) CanUnlock(address string) bool {
  return in.Sig == address
}

func (out *TxOutput) CanBeUnlocked(address string) bool {
  return out.PubKey == address
}
