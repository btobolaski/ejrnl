package workflows

import (
	"testing"
	"time"
)

var exampleEntry = `date: 2016-12-24T00:32:58Z
tags:
- Test
---
This is the body`

func TestRead(t *testing.T) {
	entry, err := read([]byte(exampleEntry))

	if err != nil {
		t.Errorf("Failed to parse the entry because %s", err)
		return
	}
	if entry.Body != "This is the body" {
		t.Error("Body didn't match expected")
	}

	if len(entry.Tags) != 1 || entry.Tags[0] != "Test" {
		t.Errorf("Tags were incorrect\ngot:      %v\nexpected: %v", entry.Tags, []string{"Test"})
	}
	if *entry.Date != time.Date(2016, 12, 24, 0, 32, 58, 0, time.UTC) {
		t.Errorf("Date didn't match expected\ngot:      %v\nexpected: %v", entry.Date, time.Date(2016, 12, 24, 0, 32, 58, 0, time.UTC))
	}
}

func TestRoundTrip(t *testing.T) {
	entry, err := read([]byte(exampleEntry))
	if err != nil {
		t.Errorf("Failed to process example because %s", err)
		return
	}
	display := format(entry)
	if display != exampleEntry {
		t.Errorf("Formatted entry doesn't match expected.\ngot:      '%s'\nexpected: '%s'", display, exampleEntry)
	}
}
