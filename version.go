// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package main

import (
	"strconv"
	"time"
)

// BuildTag placeholder
var BuildTag = "v0.0.0"

// BuildCommit placeholder
var BuildCommit = "<buildcommit>"

// buildDate placeholder
var buildDate = "0"

// BuildDate timestamp
var BuildDate = time.Unix(func() int64 { i, _ := strconv.ParseInt(buildDate, 0, 64); return i }(), 0)
