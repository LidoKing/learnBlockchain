package main

import (
  "fmt"
  "log"
  "bytes"
  "encoding/binary"
)

func ToHex(num int64) []byte {
  buff := new(bytes.Buffer)
  err := binary.Write(buff, binary.BigEndian, num)
  if err != nil {
    log.Panic(err)
  }
  return buff.Bytes()
}

func main() {
  fmt.Println(ToHex(65536))
}
