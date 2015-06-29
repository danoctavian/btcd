package ldb

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/opt"
)

type LdbBlockStore struct {
	lBatch *leveldb.Batch
	lDb *leveldb.DB
	ro  *opt.ReadOptions
}


func (bs *LdbBlockStore) PutBlock(blkKey, blkVal []byte) {
	bs.lBatch.Put(blkKey, blkVal)
}

func (bs *LdbBlockStore) GetBlock(blkHeight int64) (rsha *wire.ShaHash, rbuf []byte, err error) {
	var blkVal []byte

	key := int64ToKey(blkHeight)

	blkVal, err = bs.lDb.Get(key, bs.ro)
	if err != nil {
		log.Tracef("failed to find height %v", blkHeight)
		return // exists ???
	}

	var sha wire.ShaHash

	sha.SetBytes(blkVal[0:32])

	blockdata := make([]byte, len(blkVal[32:]))
	copy(blockdata[:], blkVal[32:])

	return &sha, blockdata, nil								
}


func (bs *LdbBlockStore) DeleteBlock(height int64) {
	bs.lBatch.Delete(int64ToKey(height))	
}




