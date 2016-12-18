package ejrnl

import (
	"time"
)

type Config struct {
	StorageDirectory, Salt string
	Pow                    uint
}

type Entry struct {
	Date     time.Time
	Body, Id string
	Tags     []string
}
