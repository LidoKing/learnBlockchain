package wallet

import (
  "crypto/ecdsa"
  "crypto/elliptic"
  "crypto/rand"
  "crypto/sha256"
  "log"
  "bytes"
  "golang.org/x/crypto/ripemd160"
)

func Handle(err error) {
  if err != nil {
    log.Panic(err)
  }
}

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
  Handle(err)

  // concatenate two slices of bytes, append(x, y...)
  public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

  return *private, public
}

func MakeWallet() *Wallet {
  private, public := NewKeyPair()
  wallet := Wallet{private, public}

  return &wallet
}

// Hash through sha256 and then ripemd160 to get public key hash
func PublicKeyHash(pubKey []byte) []byte {
  pubHash :=sha256.Sum256(pubKey)

  // Create hasher
  hasher := ripemd160.New()
  // Write pubKey into hasher
  _, err := hasher.Write(pubHash[:])
  Handle(err)

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
   pubKeyHash := PublicKeyHash(w.PublicKey)

   versionedHash := append([]byte{version}, pubKeyHash...)
   checksum := Checksum(versionedHash)

   fullHash := append(versionedHash, checksum...)
   address := Base58Encode(fullHash)

   return address
}

// Validation process:
// 1. Decode address back to full hash
// 2. Separate version public key hash and checksum
// 3. Create new checksum with public key hash and same version
// 4. Compare new cheksum and original checksum
func ValidateAddress(address string) bool {
  fullHash := Base58Decode([]byte(address))
  actualChecksum := fullHash[len(fullHash)-checksumLength:]
  version := fullHash[0]
  pubKeyHash := fullHash[1:len(fullHash)-checksumLength]
  targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

  return bytes.Compare(actualChecksum, targetChecksum) == 0
}
