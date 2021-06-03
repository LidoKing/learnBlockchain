

import (
  "log"
  "bytes"
  "encoding/gob"
)

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

/*---------------------------block---------------------------*/
type Block struct {
  Hash     []byte
  Data     []byte
  PrevHash []byte
  Nonce    int
}

func CreateBlock(data string, prevHash []byte) *Block {
  block := &Block{[]byte{}, []byte(data), prevHash, 0}
  // ^ Create block with only data and hash of previous block, other fields (hash, nonce) empty
  pow := NewProofOfWork(block)
  nonce, hash := pow.Run()
  // ^ Get noncce and hash of block after mined

  block.Hash = hash[:]
  block.Nonce = nonce

  return block
}

func Genesis() *Block {
  return CreateBlock("Genesis", []byte{})
  // ^ Create genesis block
}
