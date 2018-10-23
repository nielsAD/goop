// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package discord

import (
	"fmt"
	"strconv"
	"strings"
)

// RelayJoinMode enum
type RelayJoinMode int32

// RelayJoins
const (
	RelayJoinsSay = 1 << iota
	RelayJoinsList

	RelayJoinsBoth = RelayJoinsSay | RelayJoinsList
)

func (r RelayJoinMode) String() string {
	var res string
	if r&RelayJoinsSay != 0 {
		res += "|say"
		r &= ^RelayJoinsSay
	}
	if r&RelayJoinsList != 0 {
		res += "|list"
		r &= ^RelayJoinsList
	}
	if r != 0 {
		res += fmt.Sprintf("|0x%02X", uint32(r))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// UnmarshalText implements encoding.TextUnmarshaler
func (r *RelayJoinMode) UnmarshalText(text []byte) error {
	var s = strings.Split(strings.ToLower(string(text)), "|")
	var t RelayJoinMode

	for _, v := range s {
		switch v {
		case "say":
			t |= RelayJoinsSay
		case "list":
			t |= RelayJoinsList
		default:
			v, err := strconv.ParseInt(v, 0, 32)
			if err != nil {
				return err
			}
			t |= RelayJoinMode(v)
		}
	}

	*r = t
	return nil
}

// MarshalText implements encoding.TextMarshaler
func (r RelayJoinMode) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}
