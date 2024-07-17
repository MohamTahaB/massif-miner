package outlog

import "github.com/MohamTahaB/massif-miner/internal/snapshot"

// Define the OutLog that will serve as an accessible JSON parsing of the massif.out log files
type OutLog struct {
	Desc      string              `json:"desc"`
	Cmd       string              `json:"cmd"`
	TimeUnit  TimeUnit            `json:"timeUnit"`
	Snapshots []snapshot.Snapshot `json:"snapshots"`
}
