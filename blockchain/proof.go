package blockchain

import (
  "bytes"
  "crypto/sha256"
  "encoding/binary"
  "fmt"
  "log"
  "math"
  "math/big"
)

const Difficulty = 12
// ^ For simplicity, mining difficulty remains constant

type ProofOfWork struct {
  Block *Block
  Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
  target := big.NewInt(1)
  target.Lsh(target, uint(256-Difficulty))
  // ^ Lsh() == << (i.e. 'target' times 2 by (256-Difficulty) times)

  pow := &ProofOfWork{b, target}

  return pow
}

// Turn an int into slice of byte with the greatest number for one index being 255 (i.e. 255 = [0 0 0 0 0 0 0 255] -> 256 = [0 0 0 0 0 0 1 0])
func ToHex(num int64) []byte {
  buff := new(bytes.Buffer)
  err := binary.Write(buff, binary.BigEndian, num)
  if err != nil {
    log.Panic(err)
  }
  return buff.Bytes()
}
