package workflows

import (
	"os"
	"testing"
	"time"

	"github.com/btobolaski/ejrnl"
	"github.com/btobolaski/ejrnl/storage"
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

func TestListing(t *testing.T) {
	conf := ejrnl.Config{
		StorageDirectory: "../workflow-listing",
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

	date := time.Date(2015, 12, 24, 0, 32, 58, 0, time.UTC)
	err = driver.Write(ejrnl.Entry{
		Id:   "1",
		Date: &date,
	})
	if err != nil {
		t.Errorf("Failed to write entry because %s", err)
		return
	}
	date = time.Date(2016, 12, 24, 0, 32, 58, 0, time.UTC)
	err = driver.Write(ejrnl.Entry{
		Id:   "2",
		Date: &date,
	})
	if err != nil {
		t.Errorf("Failed to write entry because %s", err)
		return
	}

	listing, sorted, err := listing(driver)
	if err != nil {
		t.Errorf("Failed to get listing because %s", err)
		return
	}
	if listing[sorted[0]] != "2" {
		t.Errorf("Sorting was incorrect %v", sorted)
	}
}

func TestRekey(t *testing.T) {
	conf := ejrnl.Config{
		StorageDirectory: "../workflow-rekey",
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

	date := time.Date(2015, 12, 24, 0, 32, 58, 0, time.UTC)
	err = driver.Write(ejrnl.Entry{
		Id:   "1",
		Date: &date,
	})
	if err != nil {
		t.Errorf("Failed to write entry because %s", err)
		return
	}
	date = time.Date(2016, 12, 24, 0, 32, 58, 0, time.UTC)
	err = driver.Write(ejrnl.Entry{
		Id:   "2",
		Date: &date,
	})
	if err != nil {
		t.Errorf("Failed to write entry because %s", err)
		return
	}

	newConfig := ejrnl.Config{
		StorageDirectory: "../workflow-new-rekey",
		Pow:              12,
	}
	newConfig.Salt = conf.Salt

	newDriver, err := storage.NewDriver(newConfig, "password")
	if _, ok := err.(*storage.NeedsInit); !ok {
		t.Errorf("Failed to create destination driver because %s", err)
		return
	}
	if err = newDriver.Init(); err != nil {
		t.Errorf("Failed to init destination driver because %s", err)
		return
	}
	defer os.RemoveAll(newConfig.StorageDirectory)

	if err = Rekey(driver, newDriver, conf.StorageDirectory, newConfig.StorageDirectory); err != nil {
		t.Errorf("Failed to rekey because %s", err)
		return
	}

	driver, err = storage.NewDriver(conf, "password")
	if err != nil {
		t.Errorf("Failed to open renamed directory because %s", err)
		return
	}

	listing, sorted, err := listing(driver)
	if err != nil {
		t.Errorf("Failed to get listing because %s", err)
		return
	}
	if listing[sorted[0]] != "2" {
		t.Errorf("Sorting was incorrect %v", sorted)
	}
}
