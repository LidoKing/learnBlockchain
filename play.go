package main

import (
  "fmt"
  "log"
  "bytes"
  "encoding/binary"
)

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


func InitNonce(nonce int) []byte {
  data := bytes.Join(
    [][]byte{
      ToHex(int64(nonce)),
      ToHex(int64(12)),
    },
    []byte{},
  )

  return data
}

func WhatIsSliceOfByte() []byte {
  // []byte{} == []

  result := [][]byte{ToHex(2), ToHex(3)}
  // ^ [[0 0 0 0 0 0 0 2] [0 0 0 0 0 0 0 3]] (byte in byte)

  joint := bytes.Join(result, []byte{})
  // Combines two dimensional slice of byte to one slice of byte
  // ^ [0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 3]

  return joint
}

func main() {
  fmt.Println(ToHex(2))
  fmt.Println(ToHex(3))
  fmt.Println(WhatIsSliceOfByte())
}
