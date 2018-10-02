// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package gateway

// AccessLevel for user
type AccessLevel int32

// Access constants
const (
	AccessOwner     AccessLevel = 1000
	AccessAdmin     AccessLevel = 300
	AccessOperator  AccessLevel = 200
	AccessWhitelist AccessLevel = 100
	AccessVoice     AccessLevel = 1
	AccessDefault   AccessLevel = 0
	AccessIgnore    AccessLevel = -1
	AccessBlacklist AccessLevel = -100
	AccessKick      AccessLevel = -200
	AccessBan       AccessLevel = -300
)
