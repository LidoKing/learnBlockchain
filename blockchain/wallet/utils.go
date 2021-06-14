package wallet

import (
  "log"

  "github.com/mr-tron/base58"
  "github.com/LidoKing/learnBlockchain/blockchain"
)

// base58 encode gives shorter output and prevents character identification issues
// zero(0), capital O(O), capital i(I), lowercase L(l)
func Base58Encode(input []byte) []byte {
  encode := base58.Encode(input)

  // typecast string to slice of bytes
  return []byte(encode)
}

func Base58Decode(input []byte) []byte {
  // typecast slice of bytes to string
  decode, err := base58.Decode(string(input[:]))
  Handle(err)

  return decode
}
