package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"sync"
	"time"

	"ejrnl"
	"ejrnl/crypto"
)

type Driver struct {
	directory string
	key       []byte
	indexLock *sync.RWMutex
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
		indexLock: &sync.RWMutex{},
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

func (d *Driver) Write(entry ejrnl.Entry) error {
	plaintext, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	cyphertext, err := crypto.Encrypt(plaintext, d.key)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/%s.cpt", d.directory, entry.Id), cyphertext, 0600)
	if err != nil {
		return err
	}

	d.indexLock.Lock()
	index, err := d.readIndex()
	if err != nil {
		d.indexLock.Unlock()
		return err
	}
	index[entry.Date] = entry.Id
	err = d.writeIndex(index)
	d.indexLock.Unlock()
	return err
}

func (d *Driver) Read(id string) (ejrnl.Entry, error) {
	bytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.cpt", d.directory, id))
	if err != nil {
		return ejrnl.Entry{}, err
	}

	plaintext, err := crypto.Decrypt(bytes, d.key)
	if err != nil {
		return ejrnl.Entry{}, err
	}

	entry := &ejrnl.Entry{}
	err = json.Unmarshal(plaintext, entry)
	return *entry, err
}

// readIndex reads the index from the disk. The caller must have at least a read lock on d.indexLock
func (d *Driver) readIndex() (map[time.Time]string, error) {
	index := make(map[time.Time]string)
	cyphertext, err := ioutil.ReadFile(fmt.Sprintf("%s/index.cpt", d.directory))
	if err != nil {
		return index, err
	}

	plaintext, err := crypto.Decrypt(cyphertext, d.key)
	if err != nil {
		return index, err
	}

	p := &index
	err = json.Unmarshal(plaintext, p)
	return *p, err
}

// writeIndex writes an updated index file. Note that the caller must have the a lock on d.indexLock
func (d *Driver) writeIndex(index map[time.Time]string) error {
	plaintext, err := json.Marshal(index)
	if err != nil {
		return err
	}

	cyphertext, err := crypto.Encrypt(plaintext, d.key)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/index.cpt", d.directory), cyphertext, 0600)
	return err
}

// Init creates the new journal
func (d *Driver) Init() error {
	if _, err := os.Stat(d.directory); os.IsNotExist(err) {
		os.MkdirAll(d.directory, 0700)
	}

	emptyIndex := make(map[time.Time]string)
	return d.writeIndex(emptyIndex)
}
