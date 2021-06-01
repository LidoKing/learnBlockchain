package main

import (
  "fmt"
  "log"
  "bytes"
  "encoding/binary"
  "math/big"
  "crypto/sha256"
  "math"
)

/*-------------------------------basics-------------------------------*/

func StringByte(msg string) []byte {
  result := []byte(msg)
  // ^ Convert string into byte array with each character corresponding to a number in ASCII (Dec)
  // Reference: http://3.bp.blogspot.com/-OCnzi-TOfKQ/VFnGx4IWnKI/AAAAAAAADIA/IlM6qC-VEWg/s1600/ascii%2Bto%2Bdec.png

  return result
}

func WhatIsSliceOfByte() []byte {
  // []byte{}: an array of numbers less than 1 byte (i.e. >=255)

  result := [][]byte{ToHex(2), ToHex(3)}
  // ^ [[0 0 0 0 0 0 0 2] [0 0 0 0 0 0 0 3]] (byte in byte)

  joint := bytes.Join(result, []byte{})
  // Combines two dimensional slice of byte to one slice of byte
  // ^ [0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 3]

  return joint
}

/*-------------------------------block.go-------------------------------*/

type Block struct {
  Hash     []byte
  Data     []byte
  PrevHash []byte
}

func HashedByte(msg string) {
  hash := sha256.Sum256(StringByte(msg))
  final := hash[:]
	fmt.Printf("%x\n" ,final)
  // ^ %x	base 16, lower-case, two characters per byte (GO Documentation)
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

/*-------------------------------proof.go-------------------------------*/

func WhatIsLeftShift() {
  target := big.NewInt(10)
  fmt.Println("old result:", target)

  target.Lsh(target, uint(2))
  // ^ Lsh() (i.e. 'target' times 2 by (256-Difficulty) times)
  // 3: [0 0 1 1] --(left shift all 1 by 2 bits)>> 12: [1 1 0 0]
  fmt.Println("new result:", target)
}

// Convert int into byte with length: 8 and greatest number is 255 for each index/element
// 255 -> [0 0 0 0 0 0 0 255], 256 [0 0 0 0 0 0 1 0]
func ToHex(num int64) []byte {
  buff := new(bytes.Buffer)
  err := binary.Write(buff, binary.BigEndian, num)
  if err != nil {
    log.Panic(err)
  }
  return buff.Bytes()
}

// Create 'data' including nonce for hashing
func InitNonce(nonce int) []byte {
  data := bytes.Join(
    [][]byte{
      ToHex(int64(nonce)),
      ToHex(int64(250)),
    },
    []byte{},
  )

  return data
}

func Run() {
  var intHash big.Int
  var hash [32]byte

  target := big.NewInt(1)
  target.Lsh(target, uint(254))
  nonce := 0
    // This is essentially an infinite loop due to how large
    // MaxInt64 is.

  for nonce < math.MaxInt64 {
    data := InitNonce(nonce)
    hash = sha256.Sum256(data)

    fmt.Printf("Hash: %x\n", hash)
    intHash.SetBytes(hash[:])
    fmt.Println(intHash.Cmp(target))

    if intHash.Cmp(target) == -1 {
    // ^ Compare big.Int, -1 if match, 1 if not match
      break
    } else {
      nonce++
    }
  }
  fmt.Println()
  fmt.Printf("nonce: %d\n", nonce)
}

func Compare() {
  //var initHash big.Int
  var hash [32]byte
  //initHash.SetBytes([)
  fmt.Println(hash[:])
}

/*-------------------------------main-------------------------------*/

func main() {
  Run()
}
