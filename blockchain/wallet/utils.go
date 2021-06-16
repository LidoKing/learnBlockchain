package wallet

import (
  "log"

  "github.com/mr-tron/base58"
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
  if err != nil {
    log.Panic(err)
  }

  return decode
}
