package ui

import (
	"testing"
	"time"
	_ "time/tzdata" // the zone database, on any machine that runs the tests
)

// The recorded journals in testdata come from a machine in Italy, and the tests check their
// times as it showed them: run them in its time zone, whatever the zone of this machine.
func TestMain(m *testing.M) {
	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		panic(err)
	}
	time.Local = loc
	m.Run()
}
