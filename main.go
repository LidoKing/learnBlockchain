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
}
