package workflows

import (
	"os"
	"testing"
	"time"

	"code.tobolaski.com/brendan/ejrnl"
	"code.tobolaski.com/brendan/ejrnl/storage"
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

func TestInit(t *testing.T) {
	conf := ejrnl.Config{
		StorageDirectory: "../workflow-init",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	driver, err := storage.NewDriver(conf, "password")
	if _, ok := err.(*storage.NeedsInit); !ok {
		t.Errorf("Expected driver to need init but got err instead: %s", err)
		return
	}

	err = Init(driver)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Errorf("Failed to init the driver because %s", err)
	}
}

func TestImport(t *testing.T) {
	conf := ejrnl.Config{
		StorageDirectory: "../workflow-import",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	driver, err := storage.NewDriver(conf, "password")
	if _, ok := err.(*storage.NeedsInit); !ok {
		t.Errorf("Expected driver to need init but got err instead: %s", err)
		return
	}

	err = Init(driver)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Errorf("Failed to init the driver because %s", err)
		return
	}

	err = Import("./import_test.md", driver)
	if err != nil {
		t.Errorf("Failed to import file because %s", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	conf := DefaultConfig()
	if conf.StorageDirectory != "~/ejrnl" {
		t.Errorf("Storage directory %s is not the default", conf.StorageDirectory)
	}
	if conf.Pow != 19 {
		t.Errorf("Incorrect default workfactor %d", conf.Pow)
	}

	if len(conf.Salt) < 64 {
		t.Error("Salt was of the wrong length")
	}
}
