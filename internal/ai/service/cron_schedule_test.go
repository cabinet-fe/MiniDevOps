package service

import (
	"testing"
	"time"
)

func TestLocScheduleRespectsTimezone(t *testing.T) {
	sched, err := cronParser.Parse("0 12 * * *")
	if err != nil {
		t.Fatal(err)
	}
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	nextSH := locSchedule{inner: sched, loc: shanghai}.Next(base)
	nextUTC := locSchedule{inner: sched, loc: time.UTC}.Next(base)
	if nextSH.Equal(nextUTC) {
		t.Fatal("timezone must shift next fire time")
	}
	if nextSH.In(shanghai).Hour() != 12 || nextSH.In(shanghai).Minute() != 0 {
		t.Fatalf("Shanghai noon expected, got %s", nextSH.In(shanghai))
	}
	if nextUTC.UTC().Hour() != 12 {
		t.Fatalf("UTC noon expected, got %s", nextUTC.UTC())
	}
	// Asia/Shanghai is UTC+8 → 12:00 CST == 04:00 UTC
	if nextSH.UTC().Hour() != 4 {
		t.Fatalf("Shanghai noon should be 04:00 UTC, got %s", nextSH.UTC())
	}
}
