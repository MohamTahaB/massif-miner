package digger

import (
	"bufio"
	"io"

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
	return nil
}
