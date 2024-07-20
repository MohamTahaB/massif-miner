package heaptree

// Define the heap tree struct to be implemented in the detailed snapshots
type HeapTree struct {
	ID                  int
	Memory              int
	Address             string
	Func                string
	FuncFullDesc        string
	HeapAllocationLeafs []*HeapTree
}

type HeapTreeDepthCtx struct {
	HTreeDepth map[int]*HeapTree
}
