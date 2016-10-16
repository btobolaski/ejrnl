package storage

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

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
		Pow:              15,
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
		Pow:              15,
	}

	_, err := driverInit(conf)
	defer os.RemoveAll(conf.StorageDirectory)
	if err != nil {
		t.Error(err)
	}
}
