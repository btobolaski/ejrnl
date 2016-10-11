package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"ejrnl"
	"ejrnl/crypto"
)

type Driver struct {
	directory string
	key       []byte
}

// NewDriver creates a new storage driver from the specified config and password.
func NewDriver(conf ejrnl.Config, password string) (*Driver, error) {
	salt, err := base64.StdEncoding.DecodeString(conf.Salt)
	if err != nil {
		return &Driver{}, fmt.Errorf("Failed to decode salt because %s", err)
	}

	key, err := crypto.GenerateKey([]byte(password), salt, conf.Pow)
	if err != nil {
		return &Driver{}, err
	}

	driver := &Driver{
		directory: conf.StorageDirectory,
		key:       key,
	}

	err = driver.checkExists()
	if err != nil {
		return driver, err
	}

	return driver, nil
}

// checkExists checks whether the journal already exists
func (d *Driver) checkExists() error {
	current, err := user.Current()
	if err != nil {
		return fmt.Errorf("Can't retrieve the current user's information because %s", err)
	}

	path := strings.Replace(d.directory, "~", current.HomeDir, -1)
	d.directory = path

	if _, err := os.Stat(fmt.Sprintf("%s/index.ejrnl", path)); os.IsNotExist(err) {
		return &NeedsInit{msg: "the index doesn't exist"}
	}

	return nil
}

// Init creates the new journal
func (d *Driver) Init() error {
	if _, err := os.Stat(d.directory); os.IsNotExist(err) {
		os.MkdirAll(d.directory, 0700)
	}

	emptyIndex := make(map[time.Time]ejrnl.Entry)
	data, err := json.Marshal(emptyIndex)
	if err != nil {
		return fmt.Errorf("Failed to encode the empty index to json because %s", err)
	}

	encrypted, err := crypto.Encrypt(data, d.key)
	if err != nil {
		return fmt.Errorf("Failed to encrypt the empty index because %s", err)
	}

	file, err := os.Create(fmt.Sprintf("%s/index.ejrnl", d.directory))
	if err != nil {
		return fmt.Errorf("Failed to create the index file because %s", err)
	}
	defer file.Close()

	_, err = file.Write(encrypted)
	if err != nil {
		return fmt.Errorf("Failed to write the fresh index file because %s", err)
	}

	return nil
}
