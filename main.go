//go.main
package main

func main() {
  type Block struct {
    Hash     []byte
    Data     []byte
    PrevHash []byte
  }

  type BlockChain struct{
    blocks []*Block
  }

  func (b *Block) DeriveHash() { // (b *Block) 'gets instance' of block struct to access the fields
    info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
    // ^ This will join our previous block's relevant info with the new blocks
    hash := sha256.Sum256(info)
    // ^ This performs the actual hashing algorithm
    b.Hash = hash[:]
  }

  func CreateBlock(data string, prevHash []byte) *Block {
    block := &Block{[]byte{}, []byte(data), prevHash}
    // ^ Create block given current data and hash of previous block
    block.DeriveHash()
    // ^ Generate hash of current block
    return block
  }

  func (chain *BlockChain) AddBlock(data string) {
    prevBlock := chain.blocks[len(chain.blocks)-1] // 'len(chain.blocks)-1 gets the index of current latest block (i.e. previous block)'
    // ^ Get old block to get hash to create new block
    new := CreateBlock(data, prevBlock.Hash)
    // ^ Create new block
    chain.blocks = append(chain.blocks, new)
    // ^ Append new block to chain
  }
}
