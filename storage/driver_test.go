package storage

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	"code.tobolaski.com/brendan/ejrnl"
)

func makeSalt(b int) string {
	bytes := make([]byte, b)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func TestDriverNeedsInit(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "../../../needs-init",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	_, err := NewDriver(conf, "password")
	if _, ok := err.(*NeedsInit); !ok {
		t.Errorf("Vault does not exist but err %#v isn't NeedsInit", err)
	}
}

func driverInit(conf ejrnl.Config) (*Driver, error) {
	d, err := NewDriver(conf, "password")
	if _, ok := err.(*NeedsInit); !ok {
		return d, fmt.Errorf("Error wasn't NeedsInit %#v", err)
	}

	err = d.Init()
	return d, err
}

func TestDriverInit(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "../../../init-test",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	_, err := driverInit(conf)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Error(err)
	}
}

func compareEntries(first, second ejrnl.Entry) bool {
	if *first.Date != *second.Date {
		return false
	}
	if first.Body != second.Body {
		return false
	}
	if first.Id != second.Id {
		return false
	}
	if len(first.Tags) != len(second.Tags) {
		return false
	}
	for i := range first.Tags {
		if first.Tags[i] != second.Tags[i] {
			return false
		}
	}
	return true
}

func TestDriverRoundtrip(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "../../../roundtrip-test",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	d, err := driverInit(conf)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Error(err)
		return
	}

	now := time.Now()
	entry := ejrnl.Entry{
		Date: &now,
		Body: "Hello",
		Id:   "1111111111111111111",
		Tags: []string{"test"},
	}
	err = d.Write(entry)
	if err != nil {
		t.Errorf("Failed to write entry because %s", err)
		return
	}
	read, err := d.Read(entry.Id)
	if err != nil {
		t.Errorf("Failed to read entry because %s", err)
		return
	}

	if !compareEntries(entry, read) {
		t.Errorf("Entries aren't equal \n%v\n%v", entry, read)
	}
}

func TestRewriteDate(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "../../../overwrite-test",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	d, err := driverInit(conf)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Error(err)
		return
	}

	now := time.Now()
	entry := ejrnl.Entry{
		Date: &now,
		Body: "Hello",
		Id:   "1111111111111111111",
		Tags: []string{"test"},
	}
	err = d.Write(entry)
	if err != nil {
		t.Errorf("Failed to write entry because %s", err)
		return
	}
	now = time.Now()
	entry.Date = &now
	err = d.Write(entry)
	if err != nil {
		t.Errorf("failed to overwrite entry because %s", err)
		return
	}

	index, err := d.readIndex()
	if err != nil {
		t.Errorf("Failed to read index because %s", err)
		return
	}

	if len(index) > 1 {
		t.Errorf("Multiple index entries for a single entry")
	}
}

func TestIndexRecovery(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "../../../recovery-test",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	d, err := driverInit(conf)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Error(err)
		return
	}

	now := time.Now()
	entry := ejrnl.Entry{
		Date: &now,
		Body: "Hello",
		Id:   "1111111111111111111",
		Tags: []string{"test"},
	}
	err = d.Write(entry)
	if err != nil {
		t.Errorf("Failed to write entry because %s", err)
		return
	}

	os.RemoveAll(fmt.Sprintf("%s/index.cpt", conf.StorageDirectory))
	d, err = driverInit(conf)
	if err != nil {
		t.Errorf("Failed to reinit driver because %s", err)
		return
	}
	index, err := d.readIndex()
	if err != nil {
		t.Errorf("Failed to read index because %s", err)
		return
	}
	if index[*entry.Date] == "" {
		t.Errorf("index did not contain previously written entry")
		return
	}
	read, err := d.Read(index[*entry.Date])
	if err != nil {
		t.Errorf("Failed to read recovered entry because %s", err)
		return
	}
	if !compareEntries(entry, read) {
		t.Errorf("Written entry is different from recovered entry:\n%v\n%v", entry, read)
	}
}

func TestDefaultDate(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "../default-date",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	d, err := driverInit(conf)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Error(err)
		return
	}

	entry := ejrnl.Entry{
		Body: "Hello",
		Id:   "1111111111111111111",
		Tags: []string{"test"},
	}
	err = d.Write(entry)
	if err != nil {
		t.Errorf("Failed to write entry because %s", err)
		return
	}
	read, err := d.Read(entry.Id)
	if err != nil {
		t.Errorf("Failed to read entry because %s", err)
		return
	}

	if read.Date == nil {
		t.Error("Date wasn't set automatically for file")
	}
}

func TestTildeExpand(t *testing.T) {
	conf := ejrnl.Config{
		StorageDirectory: "~/tilde-expand",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	driver, _ := NewDriver(conf, "password")
	if driver.directory[:1] == "~" {
		t.Error("Didn't expand ~ in the storage path")
	}
}
