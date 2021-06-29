package blockchain

import (
  "log"
  "bytes"
  "encoding/gob"
)

type Block struct {
  Hash          []byte
  Transactions  []*Transaction
  PrevHash      []byte
  Nonce         int
}

/*---------------------------utils---------------------------*/

func Handle(err error) {
  if err != nil {
    log.Panic(err)
  }
}

// Block to byte
func (b *Block) Serialize() []byte {
  var res bytes.Buffer
  encoder := gob.NewEncoder(&res)

  err := encoder.Encode(b)

  Handle(err)

  return res.Bytes()
}

// Byte to block
func Deserialize(data []byte) *Block {
  var block Block

  decoder := gob.NewDecoder(bytes.NewReader(data))

  err := decoder.Decode(&block)

  Handle(err)

  return &block
}

/*---------------------------main---------------------------*/

func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
  // Create block with only data and hash of previous block
  // other fields (hash, nonce) empty
  block := &Block{[]byte{}, txs, prevHash, 0}
  pow := NewProofOfWork(block)

  // Get noncce and hash of block after mined
  nonce, hash := pow.Run()

  block.Hash = hash[:]
  block.Nonce = nonce

  return block
}

func (b *Block) SerializeTransactions() []byte {
  // Array of transaction IDs
  var serializedTXs [][]byte

  for _, tx := range b.Transactions {
    serializedTXs = append(serializedTXs, tx.Serialize())
  }

  tree := NewMerkleTree(serializedTXs)

  return tree.RootNode.Data
}

func Genesis(coinbase *Transaction) *Block {
  return CreateBlock([]*Transaction{coinbase}, []byte{})
}
