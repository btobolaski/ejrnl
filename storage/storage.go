package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"regexp"
	"strings"
	"sync"
	"time"

	"code.tobolaski.com/brendan/ejrnl"
	"code.tobolaski.com/brendan/ejrnl/crypto"
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
	if entry.Date == nil {
		now := time.Now()
		entry.Date = &now
	}
	plaintext, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	cyphertext, err := crypto.Encrypt(plaintext, d.key)
	if err != nil {
		return err
	}

	var previousDate time.Time

	if _, err = os.Stat(fmt.Sprintf("%s/%s.cpt", d.directory, entry.Id)); err == nil {
		old, err := d.Read(entry.Id)
		if err != nil {
			log.Printf("Failed to read previous entry because %s", err)
		} else {
			previousDate = *old.Date
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/%s.cpt", d.directory, entry.Id), cyphertext, 0600)
	if err != nil {
		return err
	}

	d.indexLock.Lock()
	defer d.indexLock.Unlock()
	index, err := d.readIndex()
	if err != nil {
		return err
	}
	delete(index, previousDate)
	index[*entry.Date] = entry.Id
	err = d.writeIndex(index)
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
	d.indexLock.Lock()
	defer d.indexLock.Unlock()
	if _, err := os.Stat(d.directory); os.IsNotExist(err) {
		os.MkdirAll(d.directory, 0700)
	}

	files, err := ioutil.ReadDir(d.directory)
	if err != nil {
		return fmt.Errorf("Failed to read directory for journal because %s")
	}
	previousEntries := []os.FileInfo{}
	for _, file := range files {
		if match, _ := regexp.MatchString("\\.cpt$", file.Name()); match {
			previousEntries = append(previousEntries, file)
		}
	}

	emptyIndex := make(map[time.Time]string)

	if len(previousEntries) > 0 {
		entryReader := make(chan *ejrnl.Entry, len(previousEntries))
		reader := func(f os.FileInfo) {
			id := strings.Replace(f.Name(), ".cpt", "", 1)
			entry, err := d.Read(id)
			if err != nil {
				log.Printf("Failed to recover %s because %s", f.Name(), err)
				entryReader <- nil
			} else {
				entryReader <- &entry
			}
		}
		for _, file := range previousEntries {
			go reader(file)
		}

		complete := 0
		failed := 0
		timer := time.NewTimer(30 * time.Second)

		for complete < len(previousEntries) {
			select {
			case entry := <-entryReader:
				complete++
				if entry == nil {
					failed++
				} else {
					emptyIndex[*entry.Date] = entry.Id
				}
			case <-timer.C:
				return errors.New("Timed out waiting for recovery to finish")
			}
		}
		if failed > 0 {
			return errors.New("Failed to recover one or more entries. See previous messages")
		}
	}

	return d.writeIndex(emptyIndex)
}
