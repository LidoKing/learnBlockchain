package wallet

import (
  "crypto/ecdsa"
  "crypto/elliptic"
  "crypto/rand"
  "crypto/sha256"
  "log"

  "golang.org/x/crypto/ripemd160"
)

const (
  checksumLength = 4
  version = byte(0x00)
)

type Wallet struct {
  PrivateKey ecdsa.PrivateKey
  PublicKey  []byte
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
  curve := elliptic.P256()

  private, err := ecdsa.GenerateKey(curve, rand.Reader)
  if err != nil {
    log.Panic(err)
  }

  // concatenate two slices of bytes, append(x, y...)
  public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

  return *private, public
}

func MakeWallet() *Wallet {
  private, public := NewKeyPair()
  wallet := Wallet{private, public}

  return &wallet
}

// Hash through sha256 and then ripemd160 to get public hash
func PublicKeyHash(pubKey []byte) []byte {
  pubHash :=sha256.Sum256(pubKey)

  // Create hasher
  hasher := ripemd160.New()
  // Write pubKey into hasher
  _, err := hasher.Write(pubHash[:])
  if err != nil {
    log.Panic(err)
  }

  // Actual hashing
  publicRipeMd := hasher.Sum(nil)
  return publicRipeMd
}

func Checksum(versionedHash []byte) []byte {
  firstHash := sha256.Sum256(versionedHash)
  secondHash := sha256.Sum256(firstHash[:])

  // Get first four bytes
  return secondHash[:checksumLength]
}

func (w Wallet) Address() []byte {
   pubHash := PublicKeyHash(w.PublicKey)

   versionedHash := append([]byte{version}, pubHash...)
   checksum := Checksum(versionedHash)

   finalHash := append(versionedHash, checksum...)
   address := Base58Encode(finalHash)

   return address
}
