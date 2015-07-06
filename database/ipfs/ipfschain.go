package ipfs

import (

	"github.com/btcsuite/btcd/wire"
	"github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	dag "github.com/ipfs/go-ipfs/merkledag"
	"github.com/btcsuite/btcutil"
	key "github.com/ipfs/go-ipfs/blocks/key"
	mh "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-multihash"
  "github.com/btcsuite/goleveldb/leveldb"
  "github.com/btcsuite/goleveldb/leveldb/opt"
  "github.com/btcsuite/btclog"

  "golang.org/x/net/context"
  "fmt"
  "bytes"
  "strconv"
  "time"
/*
	"encoding/binary"
	"os"
	"strconv"
	"sync"

	"github.com/btcsuite/btcd/database"
	"github.com/btcsuite/btclog"

	*/
)

/*
Stores block data in IPFS.
stores an index (btc hash -> ipfschain hash) in LevelDB. 

*/

// FIXME: implement
type IpfsChain struct {
	node *core.IpfsNode

	// level db
	lDb *leveldb.DB
	lBatch *leveldb.Batch
	ro  *opt.ReadOptions
	wo  *opt.WriteOptions

	// ipfs

	emptyNode *dag.Node
}

func NewIpfsChain(lDb *leveldb.DB, lBatch *leveldb.Batch, ro *opt.ReadOptions, wo *opt.WriteOptions) *IpfsChain {

	fmt.Println("###### Running an ipfs node")
	/* run ipfs node */
  r, err := fsrepo.Open("~/.ipfs")
  if err != nil {
    panic("failed to open ipfs repo")
  }

  nb := core.NewNodeBuilder().Online()
  nb.SetRepo(r)

  ctx, _ := context.WithCancel(context.Background())
  //defer cancel()

  node, err := nb.Build(ctx)
  if err != nil {
    panic("failed to build an ipfs node")
  }

  /*
  	a pre-genesis dummy 
  */
  emptyNode := &dag.Node{Data: ([]byte)("i am a dummy.")}

  node.DAG.Add(emptyNode)

	return &IpfsChain{node, lDb, lBatch, ro, wo, emptyNode}
}

func (ic IpfsChain) PutBlock(blkHeight int64, sha, prevSha *wire.ShaHash, buf []byte) {
	
	block, _ := btcutil.NewBlockFromBytes(buf)

	txsNode, _ := txsToNode(block.MsgBlock().Transactions)

  var w bytes.Buffer
	block.MsgBlock().Header.Serialize(&w)

  // store the header data only in the root dagNode
  dagNode := &dag.Node{Data: w.Bytes()}

//  fmt.Printf(" the txsNode node is %s", txsNode)
  dagNode.AddNodeLink(transactionsLink, txsNode)

	// FIXME: try with this; it probably fails because of the dummy link
	//ic.node.DAG.AddRecursive(dagNode)

	if (blkHeight == 0) { // the genesis block, there is no previous

		//err := dagNode.AddRawLink(prevBlockLink, preGenesisDummyLink)
		err := dagNode.AddNodeLink(prevBlockLink, ic.emptyNode)
		if (err != nil) {
			return
		}

	} else { // it's not the first block
		prevNodeKeyBytes, _ := ic.lDb.Get(btcHashToKey(prevSha), ic.ro)

		prevNodeKey := ipfsKeyFromBytes(prevNodeKeyBytes)

//		fmt.Printf("the prev node key is  %s", prevNodeKey)

		ctx, _ := context.WithTimeout(context.Background(), time.Second * 30) 
		prevNode, _ := ic.node.DAG.Get(ctx, prevNodeKey)

		dagNode.AddNodeLink(prevBlockLink, prevNode)
	}

	/*
		add the children and then the parent node
	*/
	ic.node.DAG.Add(txsNode)
  headerKey, _ := ic.node.DAG.Add(dagNode)

	blkKey := int64ToKey(blkHeight)
	ic.lBatch.Put(blkKey, ipfsKeyToBytes(headerKey))

	// map btc hash to the key of the object in ipfs
	ic.lBatch.Put(btcHashToKey(block.Sha()), ipfsKeyToBytes(headerKey))
}

func (ic IpfsChain) GetBlock(blkHeight int64) (rsha *wire.ShaHash, rbuf []byte, err error) {
	key := int64ToKey(blkHeight)
	ipfsKeyBytes, err := ic.lDb.Get(key, ic.ro)

	if (err != nil) {
		fmt.Printf("fetching the ipfs key failed %v", err)
		return
	}

	ipfsKey := ipfsKeyFromBytes(ipfsKeyBytes)

	ctx, _ := context.WithTimeout(context.Background(), time.Second * 30)

	dagNode, err := ic.node.DAG.Get(ctx, ipfsKey)
	if (err != nil) {
		fmt.Printf("reading dag node failed %v", err)
		return
	}

	txsNodeLink, err := dagNode.GetNodeLink(transactionsLink)
	if (err != nil) {return}

	txsNode, err := txsNodeLink.GetNode(ctx, ic.node.DAG)
	if (err != nil) {
		return
	}

  //parsing the whole 
	r := bytes.NewReader(append(dagNode.Data, txsNode.Data...))
	msgBlock := &wire.MsgBlock{}

	err = msgBlock.BtcDecode(r, pver)
	if (err != nil) {
		return
	}

	block := btcutil.NewBlock(msgBlock)
	blockBytes, err := block.Bytes()
	if (err != nil) {return}

	return block.Sha(), blockBytes, nil								
}

func (ic IpfsChain) DeleteBlock(height int64) {
	// TODO: do something about the blocks in ipfs
	ic.lBatch.Delete(int64ToKey(height))
}


func txsToNode(txs []*wire.MsgTx) (dagNode *dag.Node, err error) {
  var w bytes.Buffer

	err = wire.WriteVarInt(&w, pver, uint64(len(txs)))
	if err != nil {
		return
	}

	for _, tx := range txs {
		err = tx.BtcEncode(&w, pver)
		if err != nil {
			return
		}
	}

  serializedTxs := w.Bytes()

	dagNode = &dag.Node{Data: serializedTxs}
	return
}

func decodeTxs(buf []byte) (transactions []*wire.MsgTx, err error) {
	r := bytes.NewReader(buf)

	txCount, err := wire.ReadVarInt(r, pver)
	if err != nil {
		return
	}

	transactions = make([]*wire.MsgTx, 0, txCount)
	for i := uint64(0); i < txCount; i++ {
		tx := wire.MsgTx{}
		err = tx.BtcDecode(r, pver)
		if err != nil {
			return
		}
		transactions = append(transactions, &tx)
	}
	return
}

/*
====================
ipfschain hash index
====================

hash of a bitcoin element mapped to the hash
of the ipfs object wrapping the bitcoin object
*/
var chainIndexPrefix = []byte("b+-")

// TODO: remove these here harcodingk
var pver uint32 = 0

func btcHashToKey(sha *wire.ShaHash) []byte {
	shaBytes := sha.Bytes()
	recordLen := len(chainIndexPrefix) + len(shaBytes) 
	record := make([]byte, recordLen, recordLen)

	copy(record[0:len(chainIndexPrefix)], chainIndexPrefix)
	copy(record[len(chainIndexPrefix):recordLen], shaBytes)

	return record
}

/*
	serializing/deserializing ipfs keys for levelDB
*/
func ipfsKeyToBytes(k key.Key) []byte {
	return ([]byte)(k.B58String())
}

func ipfsKeyFromBytes(buf []byte) key.Key {
	return key.B58KeyDecode(string(buf))
}

var transactionsLink = "transactions"

var prevBlockLink = "prevBlock"

var emptyLink, _ = mh.FromHexString("QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n")

var log = btclog.Disabled

var preGenesisDummyLink = &dag.Link{Name: prevBlockLink,
																Size: 0,
																Hash: emptyLink,
																Node: nil}

// dup of the function in leveldb
func int64ToKey(keyint int64) []byte {
	key := strconv.FormatInt(keyint, 10)
	return []byte(key)
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
