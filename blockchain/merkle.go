package blockchain

import "crypto/sha256"

// Merkle tree:
//                    root
//         node                   node
//       (tx 1+2)               (tx 3+4)
//  node         node      node          node
// (tx1)        (tx2)      (tx3)         (tx4)
//                      .
//                      .
//                      .

type MerkleTree struct {
  RootNode *MerkleNode
}

type MerkleNode struct {
  Left  *MerkleNode
  Right *MerkleNode
  Data  []byte // Serialized tx
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
  node := MerkleNode{}

  // For base layer only
  if left == nil && right == nil {
    hash := sha256.Sum256(data)
    node.Data = hash[:]
  } else { // For subsequent upper layers
    prevHashes := append(left.Data, right.Data...)
    hash := sha256.Sum256(prevHashes)
    node.Data = hash [:]
  }

  node.Left = left
  node.Right = right

  return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
  var nodes []MerkleNode

  // Check if number of data is even
  if len(data) % 2 != 0 {
    // Make a copy of last datum
    data = append(data, data[len(data) - 1])
  }

  // Create base layer
  for _, datum := range data {
    node := NewMerkleNode(nil, nil, datum)
    nodes = append(nodes, *node)
  }

  // len(data) gets number of upper layers to be built including root
  for i := 0; i < len(data) / 2; i++ {
    var layer []MerkleNode

    if len(nodes) % 2 != 0 {
      nodes = append(nodes, nodes[len(nodes) - 1])
    }

    // 1st loop: tx1+tx2, 2nd loop: tx3+tx4, 3rd loop: tx5+tx6, ...
    for j := 0; j < len(nodes); j += 2 {
      node := NewMerkleNode(&nodes[j], &nodes[j + 1], nil)
      layer = append(layer, *node)
    }

    nodes = layer
  }

  // nodes[0] gives merkle root
  tree := MerkleTree{&nodes[0]}

  return &tree
}
