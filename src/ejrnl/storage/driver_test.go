package storage

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	"ejrnl"
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
	if first.Date != second.Date {
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

	entry := ejrnl.Entry{
		Date: time.Now(),
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
