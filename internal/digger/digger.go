package digger

import (
	"bufio"
	"fmt"
	"io"
	"regexp"

	"github.com/MohamTahaB/massif-miner/internal/outlog"
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
