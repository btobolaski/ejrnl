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

	"gopkg.in/yaml.v2"

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

// Import imports the specified file. It must be in the user format
func Import(path string, driver ejrnl.Driver) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	entry, err := read(data)
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
		Salt:             makeSalt(64),
		Pow:              19,
	}
}

func Print(driver ejrnl.Driver, count int) error {
	listing, sorted, err := listing(driver)
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
		fmt.Printf("%s\n\n-------------------------------------\n\n", format(entry))
	}
	return nil
}

// ListEntries outputs the date and id of the most recent count of entries. If count <= 0, it
// outputs all of the entries
func ListEntries(driver ejrnl.Driver, count int) error {
	index, sorted, err := listing(driver)
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

func NewEntry(driver ejrnl.Driver) error {
	date := time.Now()
	entry := ejrnl.Entry{Date: &date}
	tempFile := strings.Replace(fmt.Sprintf("%s/%s.ejrnl", os.TempDir(), date), " ", "-", -1)
	err := ioutil.WriteFile(tempFile, []byte(format(entry)), 0600)
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
	readEntry, err := read(bytes)
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

// listing gets a listing of all of the entries and sorts the listing's index by reverse
// chronological order
func listing(driver ejrnl.Driver) (map[time.Time]string, []time.Time, error) {
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
func read(raw []byte) (ejrnl.Entry, error) {
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
func format(entry ejrnl.Entry) string {
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
