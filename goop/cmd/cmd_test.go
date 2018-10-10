// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd_test

import (
	"testing"

	"github.com/nielsAD/goop/goop"
	"github.com/nielsAD/goop/goop/cmd"
)

func TestInterface(t *testing.T) {
	var g = goop.New(nil)
	var c cmd.Commands
	c.AddTo(g)
}
