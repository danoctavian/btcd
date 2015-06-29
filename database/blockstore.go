package database

import (
	"github.com/btcsuite/btcd/wire"
)

type BlockStore interface {
	PutBlock(blkKey, blkVal []byte)
	GetBlock(blkHeight int64) (rsha *wire.ShaHash, rbuf []byte, err error)
	DeleteBlock(height int64)
}