package digger

import (
	"strings"
	"testing"

	"github.com/MohamTahaB/massif-miner/internal/outlog"
)

func TestInitDiggerSite_OK(t *testing.T) {
	diggerSite := InitDiggerSite(strings.NewReader("test \n complete"))

	// check if diggersite scanner is nil
	if diggerSite.Scanner == nil {
		t.Fatal("digger site scanner is nil")
	}
}

func TestScan_OK(t *testing.T) {

	// Init a digger site
	diggerSite := InitDiggerSite(strings.NewReader("test \n complete"))
	// Scan for a first time
	if !diggerSite.Scan() {
		t.Fatal("first scan failed")
	}
	// Scan for a second time
	if !diggerSite.Scan() {
		t.Fatal("second scan failed")
	}

	// Final scan should be false as it is in the EOF
	if diggerSite.Scan() || diggerSite.Scanner.Err() != nil {
		t.Fatal("digger site is not at EOF or some other error occured")
	}

}

func TestText_OK(t *testing.T) {

	// Init a digger site
	diggerSite := InitDiggerSite(strings.NewReader("test \n complete"))

	// Check the text in the first token and that the scanner is not at EOF
	diggerSite.Scan()
	output := diggerSite.Text()
	diggerSite.Scan()

	if output != "test " || diggerSite.Scanner.Err() != nil {
		t.Fatal("incorrect output or unexpected scanning error")
	}
}

func TestMetaData_OK(t *testing.T) {

	// Init a digger site with a valid metadata header
	diggerSite := InitDiggerSite(strings.NewReader("desc: --massif.out\ncmd: ./file/path\ntime_unit: i\n"))

	outLog := outlog.OutLog{}

	// Check whether the output is nil
	if err := diggerSite.MetaData(&outLog); err != nil {
		t.Fatalf("metadata test error: %v", err)
	}

	// Check metadata
	if outLog.Desc != "--massif.out" {
		t.Fatalf("metadata error: incorrect desc, expected --massif.out, found %s", outLog.Desc)
	}
	if outLog.Cmd != "./file/path" {
		t.Fatalf("metadata error: incorrect cmd, expected ./file/path, found %s", outLog.Cmd)
	}
	if outLog.TimeUnit != outlog.I {
		t.Fatalf("metadata error: incorrect time unit, expected i")
	}
}

// TODO : make a test suite for all plausible metadata errors
func TestMetaData_ERR(t *testing.T) {

}

// TODO!
func TestMetaDataOnMassifLog_OK(t *testing.T) {

}
