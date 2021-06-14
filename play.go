package main

import (
  "fmt"
  "log"
  "bytes"
  "encoding/binary"
  "encoding/gob"
  "math"
  "math/big"
  "crypto/sha256"
  "reflect"
  "crypto/elliptic"
  "crypto/ecdsa"
  "crypto/rand"
  "golang.org/x/crypto/ripemd160"
)

/*-------------------------------basics-------------------------------*/

func StringByte(msg string) []byte {
  // Convert string into byte array
  // with each character corresponding to a number in ASCII (Dec)
  result := []byte(msg)

  return result
}

func WhatIsSliceOfByte() []byte {
  // []byte{}: an array of numbers each less than 1 byte (i.e. <=255)

  // [[0 0 0 0 0 0 0 2] [0 0 0 0 0 0 0 3]] (byte in byte)
  result := [][]byte{ToHex(2), ToHex(3)}

  // Combines two dimensional slice of byte to one slice of byte
  // [0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 3]
  joint := bytes.Join(result, []byte{})

  return joint
}

func (b *Block) StructToByte()  {
  var encoded bytes.Buffer
  fmt.Println(encoded)
  fmt.Println(encoded.Bytes())

  var hash [32]byte

  // Include block info before hashing
  encode := gob.NewEncoder(&encoded)
  err := encode.Encode(b)
  Handle(err)

  // Hashing function
  hash = sha256.Sum256(encoded.Bytes())
  fmt.Printf("%x", hash[:])
}

/*-------------------------------block.go-------------------------------*/

// Error handler
func Handle(err error) {
  if err != nil {
    log.Panic(err)
  }
}

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
  // ^ Join previous block's relevant info with the new block
  hash := sha256.Sum256(info)
  // ^ Performs the actual hashing algorithm
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
  Handle(err)

  return buff.Bytes()
}

const Difficulty = 25

// Create 'data' including nonce for hashing
func InitNonce(nonce int) []byte {
  data := bytes.Join(
    [][]byte{
      ToHex(int64(nonce)),
      ToHex(int64(256-Difficulty)),
    },
    []byte{},
  )

  return data
}

var Nonce int

func Run() {
  var intHash big.Int
  var hash [32]byte

  target := big.NewInt(1)
  target.Lsh(target, uint(256-Difficulty))
  nonce := 0
    // This is essentially an infinite loop due to how large
    // MaxInt64 is.

  for nonce < math.MaxInt64 {
    data := InitNonce(nonce)
    hash = sha256.Sum256(data)

    fmt.Printf("Hash: %x\n", hash)
    fmt.Println(nonce)
    intHash.SetBytes(hash[:])
    //int(intHash.Cmp(target))

    if intHash.Cmp(target) == -1 {
    // ^ Compare big.Int, -1 if match, 1 if not match
      break
    } else {
      nonce++
    }
  }
  fmt.Println()
  fmt.Printf("successful nonce: %d\n", nonce)
  Nonce = nonce
}

func Compare() {
  var intHash big.Int
  var hash [32]byte
  fmt.Println(reflect.TypeOf(hash))

  // SetBytes() turns byte into big integer
  test := intHash.SetBytes(hash[:])
  fmt.Println(reflect.TypeOf(test))
}


func Validate() bool {
  var intHash big.Int

  data := InitNonce(Nonce)

  hash := sha256.Sum256(data)
  intHash.SetBytes(hash[:])

  target := big.NewInt(1)
  target.Lsh(target, uint(256-Difficulty))

  return intHash.Cmp(target) == -1
}

/*-------------------------------wallet.go-------------------------------*/

func NewKeyPair() []byte {
  curve := elliptic.P256()

  private, err := ecdsa.GenerateKey(curve, rand.Reader)
  Handle(err)

  part1 := private.PublicKey.X.Bytes()
  part2 := private.PublicKey.Y.Bytes()

  public := append(part1, part2...)

  fmt.Printf("Part 1: %x\n", part1)
  fmt.Printf("Part 2: %x\n", part2)
  fmt.Printf("Public Key: %x\n", public)

  return public
}

func Ripemd160(pubKey []byte) {
  pubHash :=sha256.Sum256(pubKey)
  fmt.Printf("pubHash: %x\n", pubHash)

  hasher := ripemd160.New()
  fmt.Printf("hasher1: %x\n", hasher)

  _, err := hasher.Write(pubHash[:])
  Handle(err)
  fmt.Printf("hasher2: %x\n", hasher)

  publicRipeMd := hasher.Sum(nil)
  fmt.Printf("publicRipeMd: %x\n", publicRipeMd)
}

/*-------------------------------main-------------------------------*/

func main() {
  result := NewKeyPair()
  Ripemd160(result)
}
