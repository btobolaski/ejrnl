package workflows

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/termie/go-shutil"
	"gopkg.in/yaml.v2"

	"github.com/btobolaski/ejrnl"
)

func MakeSalt(b int) string {
	bytes := make([]byte, b)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

// Import imports the specified file. It must be in the user format
func Import(path string, driver ejrnl.Driver) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	entry, err := Read(data)
	if err != nil {
		return err
	}
	return driver.Write(entry)
}

// Init inits the specified driver
func Init(driver ejrnl.Driver) error {
	return driver.Init()
}

// DefaultConfig returns the default configuration file.
func DefaultConfig() ejrnl.Config {
	return ejrnl.Config{
		StorageDirectory: "~/ejrnl",
		Salt:             MakeSalt(64),
		Pow:              19,
	}
}

func Print(driver ejrnl.Driver, count int) error {
	listing, sorted, err := Listing(driver)
	if err != nil {
		return err
	}
	if count <= 0 {
		count = len(sorted)
	} else if count > len(sorted) {
		count = len(sorted)
	}

	for i := 0; i < count; i++ {
		entry, err := driver.Read(listing[sorted[i]])
		if err != nil {
			return err
		}
		fmt.Printf("%s\n\n-------------------------------------\n\n", Format(entry))
	}
	return nil
}

// ListEntries outputs the date and id of the most recent count of entries. If count <= 0, it
// outputs all of the entries
func ListEntries(driver ejrnl.Driver, count int) error {
	index, sorted, err := Listing(driver)
	if err != nil {
		return err
	}

	if count <= 0 || count > len(sorted) {
		count = len(sorted)
	}

	for i := 0; i < count; i++ {
		fmt.Printf("%s - %s\n", sorted[i], index[sorted[i]])
	}
	return nil
}

// NewEntry creates a new entry in the expected format and then opens the user's editor for them to
// edit the entry.
func NewEntry(driver ejrnl.Driver, tempDir string) error {
	date := time.Now()
	entry := ejrnl.Entry{Date: &date}
	tempFile := strings.Replace(fmt.Sprintf("%s/%s.ejrnl", tempDir, date), " ", "-", -1)
	err := ioutil.WriteFile(tempFile, []byte(Format(entry)), 0600)
	if err != nil {
		return err
	}
	err = editFile(tempFile)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadFile(tempFile)
	if err != nil {
		return err
	}
	readEntry, err := Read(bytes)
	if err != nil {
		return err
	}
	if readEntry.Body == entry.Body && len(readEntry.Tags) == 0 && readEntry.Id == "" {
		println("entry wasn't changed, not adding it to the journal")
		return os.Remove(tempFile)
	}
	err = driver.Write(readEntry)
	if err != nil {
		return err
	}
	return os.Remove(tempFile)
}

// EditEntry decrypts the specified entry and then opens the user's editor and saves the edits
func EditEntry(driver ejrnl.Driver, id, tempDir string) error {
	entry, err := driver.Read(id)
	if err != nil {
		return err
	}

	tempFile := strings.Replace(fmt.Sprintf("%s/%s.ejrnl", tempDir, entry.Id), " ", "-", -1)
	err = ioutil.WriteFile(tempFile, []byte(Format(entry)), 0600)
	if err != nil {
		return err
	}
	err = editFile(tempFile)
	if err != nil {
		fmt.Printf("an error was encountered, the editted file still exists at %s\n", tempFile)
		return err
	}
	bytes, err := ioutil.ReadFile(tempFile)
	if err != nil {
		fmt.Printf("an error was encountered, the editted file still exists at %s\n", tempFile)
		return err
	}
	entry, err = Read(bytes)
	if err != nil {
		fmt.Printf("an error was encountered, the editted file still exists at %s\n", tempFile)
		return err
	}
	err = driver.Write(entry)
	if err != nil {
		fmt.Printf("an error was encountered, the editted file still exists at %s\n", tempFile)
		return err
	}
	err = os.Remove(tempFile)
	if err != nil {
		fmt.Printf("an error was encountered, the editted file still exists at %s\n", tempFile)
	}
	return err
}

func Rekey(oldDriver, newDriver ejrnl.Driver, journalDir, tempDir string) error {
	toTransfer, err := oldDriver.List()
	if err != nil {
		return err
	}

	for _, id := range toTransfer {
		entry, err := oldDriver.Read(id)
		if err != nil {
			return err
		}
		err = newDriver.Write(entry)
		if err != nil {
			return err
		}
	}

	if err = os.RemoveAll(journalDir); err != nil {
		fmt.Printf("Failed to remove old journal directory, %s. Rekeyed journal is at %s", journalDir, tempDir)
		return err
	}
	err = shutil.CopyTree(tempDir, journalDir, nil)
	if err != nil {
		fmt.Printf("Failed to rename %s to %s. Journal is in unknown state.\n", tempDir, journalDir)
	}
	return err
}

type timeSlice []time.Time

func (ts timeSlice) Len() int {
	return len(ts)
}

func (ts timeSlice) Less(i, j int) bool {
	return ts[i].Before(ts[j])
}

func (ts timeSlice) Swap(i, j int) {
	temp := ts[i]
	ts[i] = ts[j]
	ts[j] = temp
}

// Listing gets a listing of all of the entries and sorts the listing's index by reverse
// chronological order
func Listing(driver ejrnl.Driver) (map[time.Time]string, []time.Time, error) {
	listing, err := driver.List()
	if err != nil {
		return listing, []time.Time{}, nil
	}
	keys := []time.Time{}
	for i := range listing {
		keys = append(keys, i)
	}
	ts := timeSlice(keys)
	sort.Sort(sort.Reverse(ts))
	return listing, []time.Time(ts), nil
}

// read reads the expected format and returns the parsed value.
func Read(raw []byte) (ejrnl.Entry, error) {
	raw = bytes.Replace(raw, []byte("\r"), []byte(""), -1)
	parts := bytes.SplitN(raw, []byte("\n---\n"), 2)
	if len(parts) != 2 {
		return ejrnl.Entry{}, errors.New("Format didn't match expected")
	}

	entry := &ejrnl.Entry{}

	err := yaml.Unmarshal(parts[0], entry)
	if err != nil {
		return *entry, err
	}
	entry.Body = string(parts[1])
	return *entry, nil
}

// format formats the specified entry in the expected format
func Format(entry ejrnl.Entry) string {
	body := entry.Body
	entry.Body = ""

	header, err := yaml.Marshal(&entry)
	if err != nil {
		// Really, this should never happen.
		panic(err)
	}

	return fmt.Sprintf("%s---\n%s", header, body)
}

func editFile(path string) error {
	command := os.ExpandEnv("$EDITOR")
	if command == "" {
		if _, err := os.Stat("/usr/bin/edit"); err == nil {
			command = "/usr/bin/edit"
		} else if os.Stat("/usr/bin/editor"); err == nil {
			command = "/usr/bin/editor"
		} else if os.Stat("/usr/bin/vim"); err == nil {
			command = "/usr/bin/vim"
		} else if os.Stat("/usr/bin/vi"); err == nil {
			command = "/usr/bin/vi"
		}
	}
	cmd := exec.Command(command, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
