package digger

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/MohamTahaB/massif-miner/internal/heaptree"
	"github.com/MohamTahaB/massif-miner/internal/outlog"
	"github.com/MohamTahaB/massif-miner/internal/snapshot"
	"github.com/MohamTahaB/massif-miner/internal/utils"
)

// A struct that wraps around a bufio scanner, in order to define member funcs and be able to add snapshots sequentially
type DiggerSite struct {
	Scanner *bufio.Scanner
}

// Initiates a digger site instance, from an io reader passed as input
func InitDiggerSite(r io.Reader) DiggerSite {
	return DiggerSite{
		Scanner: bufio.NewScanner(r),
	}
}

// Advances the digger site scanner to the next token.
func (dg *DiggerSite) Scan() bool {
	return dg.Scanner.Scan()
}

// Returns the most recent digger site token as a newly allocated string
func (dg *DiggerSite) Text() string {
	return dg.Scanner.Text()
}

// Sets the digger site meta data in the input outlog. Outputs a non nil error when the digger site token does not conform to the meta data section of a massif log file
func (dg *DiggerSite) MetaData(log *outlog.OutLog) error {

	// Get the description
	if !dg.Scan() {
		return fmt.Errorf("metadata error: scan not fulfilled")
	}

	// Edit the desc, cmd and time unit in the outlog, only when all of them are found
	var desc, cmd string
	var timeUnit outlog.TimeUnit

	// init meta data regex
	descRegex := regexp.MustCompile(`^desc: (.*)$`)
	cmdRegex := regexp.MustCompile(`^cmd: (.*)$`)
	timeUnitRegex := regexp.MustCompile(`^time_unit: (.*)$`)

	// Get the desc submatches, and check whether a desc is found
	descMatches := descRegex.FindStringSubmatch(dg.Text())

	if len(descMatches) < 2 {
		return fmt.Errorf("metadata error: desc not found, desc matches size is %d and the text passed to it is %s", len(descMatches), dg.Text())
	}
	desc = descMatches[1]
	// Advance the digger site to the following token
	if !dg.Scan() {
		return fmt.Errorf("metadata error: scan not fulfilled")
	}

	// Get the cmd submatches, and check whether a cmd is found
	cmdMatches := cmdRegex.FindStringSubmatch(dg.Text())
	if len(cmdMatches) < 2 {
		return fmt.Errorf("metadata error: cmd not found")
	}
	cmd = cmdMatches[1]
	// Advance the digger site to the following token
	if !dg.Scan() {
		return fmt.Errorf("metadata error: scan not fulfilled")
	}

	// Get the time unit submatches, and check whether a time unit is found
	timeUnitMatches := timeUnitRegex.FindStringSubmatch(dg.Text())
	if len(timeUnitMatches) < 2 {
		return fmt.Errorf("metadata error: time unit not found")
	}
	// Set time unit value
	switch timeUnitMatches[1] {
	case "i":
		timeUnit = outlog.I
	case "B":
		timeUnit = outlog.B
	case "ms":
		timeUnit = outlog.MS
	case "auto":
		timeUnit = outlog.AUTO
	default:
		return fmt.Errorf("metadata error: time unit not supported")
	}

	log.Cmd = cmd
	log.Desc = desc
	log.TimeUnit = timeUnit

	return nil
}

// Fetches info related to the snapshot chunk the scanner token is supposed to be at
func (dg *DiggerSite) FetchSnapshot(log *outlog.OutLog) error {

	var ss snapshot.Snapshot
	delimiter := "#-----------"

	// Handle scanning issues
	if !dg.Scan() {
		return fmt.Errorf("snapshot error: could not scan when fetching snapshot")
	}

	// a delimiter is expected
	if dg.Text() != delimiter {
		return fmt.Errorf("snapshot error: a delimiter is expected at the beginning of the snapshot")
	}

	// Handle scanning issues
	if !dg.Scan() {
		return fmt.Errorf("snapshot error: could not scan when fetching snapshot")
	}

	// Expect a line of the form "snapshot=id"
	var snapshotIDStr string
	var err error

	if snapshotIDStr, err = utils.ExtractValueOf("snapshot", dg.Text(), true); err != nil {
		return fmt.Errorf("snapshot error: %v", err)
	}

	if ss.Id, err = strconv.Atoi(snapshotIDStr); err != nil {
		return fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if !dg.Scan() {
		return fmt.Errorf("snapshot error: could not scan when fetching snapshot")
	}

	// a delimiter is expected
	if dg.Text() != delimiter {
		return fmt.Errorf("snapshot error: a delimiter is expected at the beginning of the snapshot")
	}

	// Handle scanning issues
	if !dg.Scan() {
		return fmt.Errorf("snapshot error: could not scan when fetching snapshot")
	}

	// Expect the time
	var timeVal string
	if timeVal, err = utils.ExtractValueOf("time", dg.Text(), true); err != nil {
		return fmt.Errorf("snapshot error: %v", err)
	}

	if ss.Time, err = strconv.Atoi(timeVal); err != nil {
		return fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if !dg.Scan() {
		return fmt.Errorf("snapshot error: could not scan when fetching snapshot")
	}

	// expect the mem_heap_B
	var memHeapVal string
	if memHeapVal, err = utils.ExtractValueOf("mem_heap_B", dg.Text(), true); err != nil {
		return fmt.Errorf("snapshot error: %v", err)
	}

	if ss.MemHeapB, err = strconv.Atoi(memHeapVal); err != nil {
		return fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if !dg.Scan() {
		return fmt.Errorf("snapshot error: could not scan when fetching snapshot")
	}

	// expect the mem_heap_B
	var memHeapExtraVal string
	if memHeapExtraVal, err = utils.ExtractValueOf("mem_heap_extra_B", dg.Text(), true); err != nil {
		return fmt.Errorf("snapshot error: %v", err)
	}

	if ss.MemHeapExtraB, err = strconv.Atoi(memHeapExtraVal); err != nil {
		return fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if !dg.Scan() {
		return fmt.Errorf("snapshot error: could not scan when fetching snapshot")
	}

	// expect the mem_heap_B
	var memStacksVal string
	if memStacksVal, err = utils.ExtractValueOf("mem_stacks_B", dg.Text(), true); err != nil {
		return fmt.Errorf("snapshot error: %v", err)
	}

	if ss.MemStacksB, err = strconv.Atoi(memStacksVal); err != nil {
		return fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if !dg.Scan() {
		return fmt.Errorf("snapshot error: could not scan when fetching snapshot")
	}

	// expect the mem_heap_B
	var heapTreeVal string
	if heapTreeVal, err = utils.ExtractValueOf("heap_tree", dg.Text(), false); err != nil {
		return fmt.Errorf("snapshot error: %v", err)
	}

	if heapTreeVal == "detailed" {
		//TODO! find a solution for heap trees
		ss.HeapTree = heaptree.HeapTree{}
	}

	return nil

}
