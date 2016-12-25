package ejrnl

import (
	"time"
)

type Config struct {
	StorageDirectory, Salt string
	Pow                    uint
}

type Entry struct {
	Date *time.Time `yaml:",omitempty"`
	Body string     `yaml:",omitempty"`
	Id   string     `yaml:",omitempty"`
	Tags []string   `yaml:",omitempty"`
}

type Driver interface {
	Write(Entry) error
	Read(string) (Entry, error)
	Init() error
}
