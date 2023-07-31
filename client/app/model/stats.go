package model

type Stats struct {
	Active    bool
	PeerCount uint64
	Pending   uint
	GasPrice  int64
	Syncing   bool
	NodeInfo  Node
	Block     *Block
}
