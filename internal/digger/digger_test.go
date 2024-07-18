package digger

import (
	"fmt"
	"os"
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
		t.Fatalf("metadata test error: incorrect desc, expected --massif.out, found %s", outLog.Desc)
	}
	if outLog.Cmd != "./file/path" {
		t.Fatalf("metadata test error: incorrect cmd, expected ./file/path, found %s", outLog.Cmd)
	}
	if outLog.TimeUnit != outlog.I {
		t.Fatalf("metadata test error: incorrect time unit, expected i")
	}
}

func TestMetaData_AllTimeUnits_OK(t *testing.T) {

	// Init the UTests struct
	type uTest struct {
		timeUnit         string
		expectedTimeUnit outlog.TimeUnit
	}

	var uTests = []uTest{
		{"i", outlog.I},
		{"B", outlog.B},
		{"ms", outlog.MS},
		{"auto", outlog.AUTO},
	}

	var ol outlog.OutLog
	var diggerSite DiggerSite
	for _, test := range uTests {
		ol = outlog.OutLog{}
		diggerSite = InitDiggerSite(strings.NewReader(fmt.Sprintf("desc: --massif.out\ncmd: ./file/path\ntime_unit: %s\n", test.timeUnit)))
		if err := diggerSite.MetaData(&ol); err != nil {
			t.Fatalf("metadata test error: %v", err)
		}

		// Check the time unit
		if ol.TimeUnit != test.expectedTimeUnit {
			t.Fatal("metadata test error: unexpected time unit")
		}
	}

}

func TestMetaDataOnMassifLog_OK(t *testing.T) {

	// Open the massif.out log in the artifacts
	file, err := os.Open("../utils/artifacts/massif.out.log")
	if err != nil {
		t.Fatalf("error opening the massif.out log: %v", err)
	}

	defer file.Close()

	// Init digger site and outlog
	dg := InitDiggerSite(file)
	ol := outlog.OutLog{}

	if err = dg.MetaData(&ol); err != nil {
		t.Fatalf("metadata test error: error reading from the massif.out: %v", err)
	}

	// Check whether the info are correct.
	// CAUTION: change in the artifacts should be taken into account here as well
	if ol.Desc != "--massif-out-file=massif.out.log" || ol.Cmd != "./alloc_dealloc" || ol.TimeUnit != outlog.I {
		t.Fatal("metadata test error: the values of the log metadata are not as expected")
	}
}

func TestSnapshotOnMassifLog_OK(t *testing.T) {

	// Open the massif.out log in the artifacts
	file, err := os.Open("../utils/artifacts/massif.out.log")
	if err != nil {
		t.Fatalf("error opening the massif.out log: %v", err)
	}

	defer file.Close()

	// Init digger site and outlog
	dg := InitDiggerSite(file)
	ol := outlog.OutLog{}

	if err = dg.MetaData(&ol); err != nil {
		t.Fatalf("snapshot test error: error reading from the massif.out: %v", err)
	}

	// Check whether the info are correct.
	// CAUTION: change in the artifacts should be taken into account here as well
	if ol.Desc != "--massif-out-file=massif.out.log" || ol.Cmd != "./alloc_dealloc" || ol.TimeUnit != outlog.I {
		t.Fatal("snapshot test error: the values of the log metadata are not as expected")
	}

	for i := 0; i < 60; i++ {
		err = dg.FetchSnapshot(&ol)
		if err != nil {
			t.Fatalf("snapshot error at iter %d: %v", i, err)
		}
	}
}
