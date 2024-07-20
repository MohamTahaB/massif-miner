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
	Scanner  *bufio.Scanner
	HTreeCtx heaptree.HeapTreeDepthCtx
}

// Initiates a digger site instance, from an io reader passed as input
func InitDiggerSite(r io.Reader) DiggerSite {
	return DiggerSite{
		Scanner: bufio.NewScanner(r),
		HTreeCtx: heaptree.HeapTreeDepthCtx{
			HTreeDepth: make(map[int]*heaptree.HeapTree),
		},
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

	delimiter := "#-----------"

	// Case when there are snapshots: the first delimiter is met

	if err := dg.AdvanceLine(); err != nil {
		if dg.Scanner.Err() != nil {
			return fmt.Errorf("metadata error: %v", dg.Scanner.Err())
		}

		// at EOF
		return nil
	}

	if dg.Text() == delimiter {
		return nil
	}

	return fmt.Errorf("metadata error: issue with first line following metadata")
}

// Advances the digger site to the following line
// Returns potential scanning errors
func (dg *DiggerSite) AdvanceLine() error {
	if !dg.Scan() {
		return fmt.Errorf("error advancing the digger site: %v", dg.Scanner.Err())
	}
	return nil
}

// Fetches info related to the snapshot chunk the scanner token is supposed to be at.
// Returns a bool (whether the digger site is at EOF), and an error, if encountered
func (dg *DiggerSite) FetchSnapshot(log *outlog.OutLog) (bool, error) {

	// TODO! consider that in case there are no snapshots, the log snapshots slice will contain an extra empty snapshot
	ss := snapshot.Snapshot{}
	delimiter := "#-----------"

	// Handle scanning issues
	if err := dg.AdvanceLine(); err != nil {
		// Two possible cases, either at EOF, or at an error that should be reported
		if dg.Scanner.Err() != nil {
			return false, fmt.Errorf("snapshot error: %v", err)
		} else {
			return true, nil
		}
	}

	// Expect a line of the form "snapshot=id"
	var snapshotIDStr string
	var err error

	if snapshotIDStr, err = utils.ExtractValueOf("snapshot", dg.Text(), true); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	if ss.Id, err = strconv.Atoi(snapshotIDStr); err != nil {
		return false, fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if err := dg.AdvanceLine(); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	// a delimiter is expected
	if dg.Text() != delimiter {
		return false, fmt.Errorf("snapshot error: a delimiter is expected at the beginning of the snapshot")
	}

	// Handle scanning issues
	if err := dg.AdvanceLine(); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	// Expect the time
	var timeVal string
	if timeVal, err = utils.ExtractValueOf("time", dg.Text(), true); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	if ss.Time, err = strconv.Atoi(timeVal); err != nil {
		return false, fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if err := dg.AdvanceLine(); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	// expect the mem_heap_B
	var memHeapVal string
	if memHeapVal, err = utils.ExtractValueOf("mem_heap_B", dg.Text(), true); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	if ss.MemHeapB, err = strconv.Atoi(memHeapVal); err != nil {
		return false, fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if err := dg.AdvanceLine(); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	// expect the mem_heap_B
	var memHeapExtraVal string
	if memHeapExtraVal, err = utils.ExtractValueOf("mem_heap_extra_B", dg.Text(), true); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	if ss.MemHeapExtraB, err = strconv.Atoi(memHeapExtraVal); err != nil {
		return false, fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if err := dg.AdvanceLine(); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	// expect the mem_heap_B
	var memStacksVal string
	if memStacksVal, err = utils.ExtractValueOf("mem_stacks_B", dg.Text(), true); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	if ss.MemStacksB, err = strconv.Atoi(memStacksVal); err != nil {
		return false, fmt.Errorf("snapshot error when converting a string: %v", err)
	}

	// Handle scanning issues
	if err := dg.AdvanceLine(); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	// expect the mem_heap_B
	var heapTreeVal string
	if heapTreeVal, err = utils.ExtractValueOf("heap_tree", dg.Text(), false); err != nil {
		return false, fmt.Errorf("snapshot error: %v", err)
	}

	ss.IsPeak = false
	var atEOF bool
	if heapTreeVal == "detailed" || heapTreeVal == "peak" {

		ss.HeapTree = &heaptree.HeapTree{}

		if heapTreeVal == "peak" {
			ss.IsPeak = true
		}

		rootRegex := regexp.MustCompile(`^n(\d+): (\d+) \(([^)]+)\)`)
		descendenceRegex := regexp.MustCompile(`^n(\d+): (\d+) ([0-9A-Fa-fx]+): (.*?) \((?:in ([^)]*)|([^)]*))\)`)
		belowThresholdRegex := regexp.MustCompile(`.*below massif's threshold.*`)

		for {
			nextLine := dg.Scan()
			// Stop if the delimiter is found, or EOF
			if atEOF = (!nextLine && dg.Scanner.Err() == nil); atEOF {
				break
			}
			if (nextLine && dg.Text() == delimiter) || atEOF {
				break
			} else if !nextLine {
				return false, fmt.Errorf("snapshot error: %v", dg.Scanner.Err())
			}

			// Check the depth of the current line of the heap tree
			htLine, depth := utils.LeadingSpaces(dg.Text())

			if belowThresholdRegex.MatchString(htLine) {
				continue
			}

			// Root of the Heap Tree
			if depth == 0 {
				dg.HTreeCtx.HTreeDepth[0] = ss.HeapTree
				match := rootRegex.FindStringSubmatch(htLine)

				// Check if the line has the root id, the mem size and the func desc
				if len(match) < 4 {
					return false, fmt.Errorf("snapshot error: unsufficient args for the root htree line: %s", htLine)
				}

				ss.HeapTree.ID, err = strconv.Atoi(match[1])
				// Handle conversion error
				if err != nil {
					return false, fmt.Errorf("snapshot error: conversion error")
				}

				ss.HeapTree.Address = "root"
				ss.HeapTree.Memory, err = strconv.Atoi(match[2])
				// Handle conversion error
				if err != nil {
					return false, fmt.Errorf("snapshot error: conversion error")
				}
				ss.HeapTree.Func = match[3]
				ss.HeapTree.FuncFullDesc = match[3]
			} else {
				// Depth is strictly positive, a node with depth n belongs to the htree decendence of the last seen leaf of depth n-1
				newHeapTreeEntry := &heaptree.HeapTree{}

				dg.HTreeCtx.HTreeDepth[depth] = newHeapTreeEntry

				match := descendenceRegex.FindStringSubmatch(htLine)

				// Check if the line has all expected info
				if len(match) < 6 {
					return false, fmt.Errorf("snapshot error: unsufficient args for the following htree line: %s, %v", htLine, match)
				}

				newHeapTreeEntry.ID, err = strconv.Atoi(match[1])
				// Handle conversion error
				if err != nil {
					return false, fmt.Errorf("snapshot error: conversion error")
				}

				newHeapTreeEntry.Memory, err = strconv.Atoi(match[2])
				// Handle conversion error
				if err != nil {
					return false, fmt.Errorf("snapshot error: conversion error")
				}
				newHeapTreeEntry.Address = match[3]
				newHeapTreeEntry.Func = match[4]
				newHeapTreeEntry.FuncFullDesc = match[5]

				// Add this node to the list of the descendences of the last seen node of depth -1
				dg.HTreeCtx.HTreeDepth[depth-1].HeapAllocationLeafs = append(dg.HTreeCtx.HTreeDepth[depth-1].HeapAllocationLeafs, newHeapTreeEntry)
			}
		}
	} else {
		nextLine := dg.Scan()
		if (!nextLine && dg.Scanner.Err() == nil) && (nextLine && dg.Text() != delimiter) {
			return false, fmt.Errorf("snapshot error: expected a delimiter or EOF")
		}
	}
	log.Snapshots = append(log.Snapshots, ss)
	return atEOF, nil

}
