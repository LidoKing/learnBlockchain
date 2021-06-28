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

// Mining difficulty remains constant for simplicity
const Difficulty = 19

type ProofOfWork struct {
  Block *Block
  Target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
  target := big.NewInt(1)
  target.Lsh(target, uint(256-Difficulty))
  // ^ Lsh() == << (i.e. 'target' times 2 by (256-Difficulty) times)

  pow := &ProofOfWork{b, target}

  return pow
}

// Convert int into slice of byte with the greatest number for one index being 255 (i.e. 255 = [0 0 0 0 0 0 0 255] -> 256 = [0 0 0 0 0 0 1 0])
func ToHex(num int64) []byte {
  buff := new(bytes.Buffer)
  err := binary.Write(buff, binary.BigEndian, num)
  if err != nil {
    log.Panic(err)
  }
  return buff.Bytes()
}

// Create data that is combined with nonce for hashing
func (pow *ProofOfWork) InitData(nonce int) []byte {
  data := bytes.Join(
    [][]byte{
      pow.Block.PrevHash,
      pow.Block.HashTransactions(),
      ToHex(int64(nonce)),
      ToHex(int64(Difficulty)),
    },
    []byte{},
  )

  return data
}

// The actual 'mining' function
func (pow *ProofOfWork) Run() (int, []byte) {
  var intHash big.Int
  var hash [32]byte

  nonce := 0

  fmt.Println()
  // This is essentially an infinite loop due to how large MaxInt64 is.
  for nonce < math.MaxInt64 {
    data := pow.InitData(nonce)
    hash = sha256.Sum256(data)

    fmt.Printf("\r%x", hash)
    intHash.SetBytes(hash[:])

    if intHash.Cmp(pow.Target) == -1 {
      break
    } else {
      nonce++
    }
  }
  fmt.Println()

  return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
  var intHash big.Int

  data := pow.InitData(pow.Block.Nonce)

  hash := sha256.Sum256(data)
  intHash.SetBytes(hash[:])

  return intHash.Cmp(pow.Target) == -1
}
