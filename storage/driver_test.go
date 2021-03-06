package storage

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/btobolaski/ejrnl"
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

func TestIncorrectPassword(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "../incorrect-password",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	_, err := driverInit(conf)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = NewDriver(conf, "incorrect-password")
	if err == nil {
		t.Error("An incorrect password didn't cause the driver to error")
	}
}

func compareEntries(first, second ejrnl.Entry) bool {
	if !(*first.Date).Equal(*second.Date) {
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
		StorageDirectory: "./roundtrip-test",
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
		StorageDirectory: "./recovery-test",
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
	index, err := d.List()
	if err != nil {
		t.Errorf("Failed to read index because %s", err)
		return
	}
	if index[now.Local()] == "" {
		log.Printf("listing: %v\ndate: %s", index, now)
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

func TestDefaultId(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "./default-id",
		Salt:             makeSalt(32),
		Pow:              12,
	}

	d, err := driverInit(conf)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Error(err)
		return
	}

	date := time.Now()
	entry := ejrnl.Entry{
		Body: "Hello",
		Date: &date,
		Tags: []string{"test"},
	}
	err = d.Write(entry)
	if err != nil {
		t.Errorf("Failed to write entry because %s", err)
		return
	}
	listing, err := d.List()
	if err != nil {
		t.Errorf("Failed to get a listing because %s", err)
		return
	}
	read, err := d.Read(listing[date.Local()])
	if err != nil {
		log.Printf("listing: %v\ndate: %s", listing, date.Local())
		t.Errorf("Failed to read entry because %s", err)
		return
	}

	if read.Id == "" {
		t.Error("ID wasn't set automatically")
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

func TestV1Decode(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "./v1-decode-test",
		Salt:             "W0qqYZBcZXo8yYudevU69F3bPblsg7zZ51hihbT+72w=",
		Pow:              12,
	}

	d, err := NewDriver(conf, "password")
	if err != nil {
		t.Errorf("Failed to create driver because %s", err)
		return
	}

	date := time.Date(2016, 12, 23, 2, 3, 4, 0, time.UTC)
	entry := ejrnl.Entry{
		Date: &date,
		Body: "Hello",
		Id:   "1111111111111111111",
		Tags: []string{"test"},
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

func TestV2Decode(t *testing.T) {
	t.Parallel()
	conf := ejrnl.Config{
		StorageDirectory: "./v2-decode-test",
		Salt:             "W0qqYZBcZXo8yYudevU69F3bPblsg7zZ51hihbT+72w=",
		Pow:              12,
	}

	d, err := NewDriver(conf, "password")
	if err != nil {
		t.Errorf("Failed to create driver because %s", err)
		return
	}

	date := time.Date(2016, 12, 23, 2, 3, 4, 0, time.UTC)
	entry := ejrnl.Entry{
		Date: &date,
		Body: "Hello",
		Id:   "1111111111111111111",
		Tags: []string{"test"},
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
