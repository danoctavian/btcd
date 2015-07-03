package ipfs

import (

	"github.com/btcsuite/btcd/wire"
	"github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	dag "github.com/ipfs/go-ipfs/merkledag"

  "golang.org/x/net/context"

/*
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/btcsuite/btcd/database"
	"github.com/btcsuite/btclog"
	"github.com/btcsuite/btcutil"
	*/
)

// FIXME: implement
type IpfsChain struct {
	node *core.IpfsNode
}

func NewIpfsChain() *IpfsChain {

	/* run ipfs node */
  r, err := fsrepo.Open("~/.ipfs")
  if err != nil {
    panic(err)
  }

  nb := core.NewNodeBuilder().Online()
  nb.SetRepo(r)

  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

  node, err := nb.Build(ctx)
  if err != nil {
    panic(err)
  }

	return &IpfsChain{node}
}

func (ic IpfsChain) PutBlock(blkKey []byte, prevSha *wire.ShaHash, blkVal []byte) {


	dagNode := &dag.Node{Data: []byte("hello my friend")}
	ic.node.DAG.Add(dagNode)
}

func (ic IpfsChain) GetBlock(blkHeight int64) (rsha *wire.ShaHash, rbuf []byte, err error) {
	return nil, nil, nil								
}

func (ic IpfsChain) DeleteBlock(height int64) {
}

/*

What to implement:

## Storage 

Store the core blockchain information in IPFS. This means block
headers and transactions.

Store the transactions as just 1 big block of data.

1 block of transactions has 1 MB. That's ok.
Since levelDB

Let LevelDB do the indexing
and all the auxilliary information holding.

### Code 

change this so it points to an IPFS hash that gives you the block
getBlkByHeight

setBlk(sha *wire.ShaHash, blkHeight int64, buf []byte)

use this to store in ipfs

tx data is read with this just fine right now.


## Sync

To sync with the vanilla network you got through the normal protocol.

To sync with the ipfschain, you just ask for the tip of the blockchain
and pin it. this will download the entire thing.

The issue with the above is that you cannot instantly confirm that the ipfs
blockchain matches the vanilla blockchain because the vanilla blockchain 
data is "wrapped" in ipfs DAG nodes.


So you cannot verify that you have the right tip of the ipfschain with a
vanilla blockchain by holding just the tip of the ipfschain and the last 
block header of the vanilla blockchain. you need to walk the blockchain
and check that the wrapped blocks chain together correctly.

(since the real links are wrapped, we don't get the benefit of veryfing
that a DAG is the same by holding it's parent)
*/
