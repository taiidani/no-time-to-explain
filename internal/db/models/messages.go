package models

import (
	"errors"
	"fmt"
	"strings"
)

func (q *Queries) ValidateMessage(m Message) error {
	var ret error

	m.Sender = strings.TrimPrefix(m.Sender, "@")
	m.Response = strings.TrimSpace(m.Response)

	if len(m.Trigger) < 4 || len(m.Response) < 4 {
		ret = errors.Join(ret, fmt.Errorf("provided inputs need to be at least 4 characters"))
	}
	return ret
}
