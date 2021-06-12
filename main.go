package main

import (
  "os"
  "github.com/LidoKing/learnBlockchain/cli"
)

func main() {
  // defer prevents corruption of bytes that are going into database
  // by making sure all actions are finished before closing database
  defer os.Exit(0)

  cmd := cli.CommandLine{}
  cmd.Run()
}
