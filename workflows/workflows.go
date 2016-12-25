package workflows

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	_ "time"

	"gopkg.in/yaml.v2"

	"code.tobolaski.com/brendan/ejrnl"
)

// Import imports the specified file
func Import(path string, driver ejrnl.Driver) error {
	return nil
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

	log.Printf("%#v", entry)

	header, err := yaml.Marshal(&entry)
	log.Printf("%s", header)
	if err != nil {
		// Really, this should never happen.
		panic(err)
	}

	return fmt.Sprintf("%s---\n%s", header, body)
}
