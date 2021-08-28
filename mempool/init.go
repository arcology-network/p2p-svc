package mempool

import "encoding/gob"

func init() {
	gob.Register([][]byte{})
}
