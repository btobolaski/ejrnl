package storage

import (
	"fmt"
)

type NeedsInit struct {
	msg string
}

func (s *NeedsInit) Error() string {
	return fmt.Sprintf("The journal needs to be inited because %s", s.msg)
}
