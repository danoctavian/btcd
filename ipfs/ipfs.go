package ipfs

import (
  "fmt"
  //"github.com/btcsuite/btcd/wire"
//  u "github.com/ipfs/go-ipfs/util"
  mh "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-multihash"
)

func TestRun() {
  fmt.Println("this is a test run")

  data := []byte("hello!")
  h, err := mh.Sum(data, mh.DSHA2_256, -1)
  fmt.Println(h.HexString())
  fmt.Println(err)
}
