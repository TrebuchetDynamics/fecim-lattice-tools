package gui

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/synctest"
)

func TestDataLoggerPreservesFiveISPPTargetsUnderBacklog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hysteresis.csv")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	synctest.Test(t, func(t *testing.T) {
		// Hold the consumer until the producer fills the queue, deterministically
		// reproducing a slow CSV writer without sleeps or host-load assumptions.
		logger := &HysteresisDataLogger{path: path, file: file, rows: make(chan HysteresisSnapshot, 1)}
		logger.Record(HysteresisSnapshot{Waveform: "ISPP", WrdTargetLevel: 1})
		done := make(chan struct{})
		go func() {
			for _, target := range []int{33, 17, 3, 2} {
				logger.Record(HysteresisSnapshot{Waveform: "ISPP", WrdTargetLevel: target})
			}
			close(done)
		}()
		synctest.Wait()

		logger.wg.Add(1)
		go logger.run()
		<-done
		if err := logger.Close(); err != nil {
			t.Fatal(err)
		}
	})

	data, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	rows, err := csv.NewReader(data).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	targetColumn := -1
	for i, name := range rows[0] {
		if name == "wrd_target_level" {
			targetColumn = i
		}
	}
	if targetColumn < 0 {
		t.Fatal("missing wrd_target_level column")
	}
	var targets []string
	for _, row := range rows[1:] {
		targets = append(targets, row[targetColumn])
	}
	if want := []string{"1", "33", "17", "3", "2"}; !reflect.DeepEqual(targets, want) {
		t.Fatalf("CSV targets = %v, want %v", targets, want)
	}
}
