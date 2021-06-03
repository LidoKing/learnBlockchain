package blockchain

type BlockChain struct{
  Blocks []*Block
}

func (chain *BlockChain) AddBlock(data string) {
  prevBlock := chain.Blocks[len(chain.Blocks)-1] // 'len(chain.blocks)-1 gets the index of current latest block (i.e. previous block)'
  // ^ Get old block to get hash to create new block
  new := CreateBlock(data, prevBlock.Hash)
  // ^ Create new block
  chain.Blocks = append(chain.Blocks, new)
  // ^ Append new block to chain
}

func InitBlockChain() *BlockChain {
  return &BlockChain{[]*Block{Genesis()}}
}
