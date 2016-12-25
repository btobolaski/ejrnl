package workflows

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"

	_ "time"

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
