// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package gateway

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// AccessLevel for user
type AccessLevel int32

// Access constants
var (
	AccessMax       = AccessLevel(math.MaxInt32)
	AccessOwner     = AccessLevel(1000)
	AccessAdmin     = AccessLevel(300)
	AccessOperator  = AccessLevel(200)
	AccessWhitelist = AccessLevel(100)
	AccessVoice     = AccessLevel(1)
	AccessDefault   = AccessLevel(0)
	AccessIgnore    = AccessLevel(-1)
	AccessKick      = AccessLevel(-100)
	AccessBan       = AccessLevel(-200)
	AccessBlacklist = AccessLevel(-300)
	AccessMin       = AccessLevel(math.MinInt32)

	ConStrings = []string{"max", "owner", "admin", "operator", "whitelist", "voice", "", "ignore", "kick", "ban", "blacklist", "min"}
	ConLevels  = []AccessLevel{AccessMax, AccessOwner, AccessAdmin, AccessOperator, AccessWhitelist, AccessVoice, AccessDefault, AccessIgnore, AccessKick, AccessBan, AccessBlacklist, AccessMin}
)

func (l AccessLevel) String() string {
	if l >= 0 {
		for i := 0; i < len(ConLevels); i++ {
			var v = ConLevels[i]
			if l == v {
				return ConStrings[i]
			}
			if l > v {
				return fmt.Sprintf("%s+%d", ConStrings[i], l-v)
			}
		}
	} else {
		for i := len(ConLevels) - 1; i >= 0; i-- {
			var v = ConLevels[i]
			if l == v {
				return ConStrings[i]
			}
			if l < v {
				return fmt.Sprintf("%s-%d", ConStrings[i], v-l)
			}
		}
	}
	return fmt.Sprintf("%d", l)
}

// MarshalText implements encoding.TextMarshaler
func (l AccessLevel) MarshalText() (text []byte, err error) {
	return []byte(l.String()), nil
}

var accessPat = regexp.MustCompile("(?i)^(" + strings.Join(ConStrings, "|") + ")([+-][0-9]+)?$")

// UnmarshalText implements encoding.TextUnmarshaler
func (l *AccessLevel) UnmarshalText(text []byte) error {
	var m = accessPat.FindSubmatch(text)

	var res AccessLevel
	var mod string
	if m != nil {
		mod = string(m[2])

		var str = string(m[1])
		var brk bool
		for i, s := range ConStrings {
			if strings.EqualFold(str, s) {
				res = ConLevels[i]
				brk = true
				break
			}
		}
		if !brk {
			return ErrUnknownAccess
		}
	} else {
		mod = string(text)
	}

	if mod != "" {
		v, err := strconv.ParseInt(string(mod), 0, 32)
		if err != nil {
			return err
		}

		res += AccessLevel(v)
	}

	*l = res
	return nil
}
