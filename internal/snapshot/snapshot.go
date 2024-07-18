package snapshot

import (
	"github.com/MohamTahaB/massif-miner/internal/heaptree"
)

// Define the snapshot struct
type Snapshot struct {
	Id            int `json:"id"`
	Time          int `json:"time"`
	MemHeapB      int `json:"memHeapB"`
	MemHeapExtraB int `json:"memHeapExtraB"`
	MemStacksB    int `json:"memStacks"`
	HeapTree      heaptree.HeapTree
	IsPeak        bool
}
