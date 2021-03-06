package wallet

import (
  "bytes"
  "crypto/elliptic"
  "encoding/gob"
  "fmt"
  "io/ioutil"
  "os"
)

const walletFile = "./tmp/wallets_%s.data"

type Wallets struct {
  Wallets map[string]*Wallet
}

func (ws *Wallets) SaveFile(nodeID string) {
  var content bytes.Buffer
  walletFile := fmt.Sprintf(walletFile, nodeID)

  gob.Register(elliptic.P256())

  enc := gob.NewEncoder(&content)
  err := enc.Encode(ws)
  Handle(err)

  // Write encoded content into designated file
  err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
  Handle(err)
}

func (ws *Wallets) LoadFile(nodeID string) error {
  walletFile := fmt.Sprintf(walletFile, nodeID)

  if _, err := os.Stat(walletFile); os.IsNotExist(err) {
    return err
  }

  var wallets Wallets

  fileContent, err := ioutil.ReadFile(walletFile)

  gob.Register(elliptic.P256())
  dec := gob.NewDecoder(bytes.NewReader(fileContent))
  err = dec.Decode(&wallets)
  if err != nil {
    return err
  }

  ws.Wallets = wallets.Wallets

  return nil
}

// Get wallets from database
func LoadWallets(nodeID string) (*Wallets, error) {
  wallets := Wallets{}
  wallets.Wallets = make(map[string]*Wallet)
  err := wallets.LoadFile(nodeID)

  return &wallets, err
}

func (ws *Wallets) AddWallet() string {
  wallet := MakeWallet()
  address := fmt.Sprintf("%s", wallet.Address())

  ws.Wallets[address] = wallet

  return address
}

func (ws Wallets) GetWallet(address string) Wallet {
  return *ws.Wallets[address]
}

func (ws *Wallets) GetAllAddresses() []string {
  var addresses []string

  for address, _ := range ws.Wallets {
    addresses = append(addresses, address)
  }

  return addresses
}
