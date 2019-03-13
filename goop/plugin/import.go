// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
	luar "github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
	"github.com/nielsAD/gowarcraft3/protocol/capi"
)

func importGlobal(ls *lua.LState) {
	for k, t := range _global {
		ls.SetGlobal(k, luar.New(ls, t))
	}
	ls.SetGlobal("events", importEvents(ls))
	ls.SetGlobal("access", importAccess(ls))
}

func importPreload(ls *lua.LState) {
	ls.PreloadModule("go.errors", preloader(_errors))
	ls.PreloadModule("go.fmt", preloader(_fmt))
	ls.PreloadModule("go.io", preloader(_io))
	ls.PreloadModule("go.context", preloader(_context))
	ls.PreloadModule("go.time", preloader(_time))
	ls.PreloadModule("go.bytes", preloader(_bytes))
	ls.PreloadModule("go.strings", preloader(_strings))
	ls.PreloadModule("go.strconv", preloader(_strconv))
	ls.PreloadModule("go.regex", preloader(_regex))
	ls.PreloadModule("go.json", preloader(_json))
	ls.PreloadModule("go.sort", preloader(_sort))
	ls.PreloadModule("go.url", preloader(_url))
	ls.PreloadModule("go.http", preloader(_http))
	ls.PreloadModule("go.websocket", preloader(_websocket))
	ls.PreloadModule("go.reflect", preloader(_reflect))
}

func preloader(m map[string]interface{}) lua.LGFunction {
	return func(ls *lua.LState) int {
		ls.Push(importMap(ls, m))
		return 1
	}
}

func importEvents(ls *lua.LState) *lua.LTable {
	var tab = ls.NewTable()
	for k, t := range _events {
		var str = strings.TrimLeft(reflect.TypeOf(t).String(), "*")
		ls.SetField(tab, k, luar.New(ls, t))
		ls.SetField(tab, str, luar.NewType(ls, t))
	}
	return tab
}

func importAccess(ls *lua.LState) *lua.LTable {
	var tab = ls.NewTable()
	for i, str := range gateway.ConStrings {
		if len(str) == 0 {
			str = "Default"
		}
		str = strings.Title(str)
		ls.SetField(tab, str, lua.LNumber(gateway.ConLevels[i]))
		ls.SetTable(tab, lua.LNumber(gateway.ConLevels[i]), lua.LString(str))
	}
	return tab
}

func importMap(ls *lua.LState, m map[string]interface{}) *lua.LTable {
	var tab = ls.NewTable()
	for k, t := range m {
		ls.SetField(tab, k, luar.New(ls, t))
	}
	return tab
}

// goop.Command wrapper for plugins
type cmdCallback func(t *gateway.Trigger, gw gateway.Gateway) error
type cmd struct{ cb cmdCallback }

func (c *cmd) CanExecute(t *gateway.Trigger) bool                                 { return true }
func (c *cmd) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error { return c.cb(t, gw) }

var _global = map[string]interface{}{
	"gotypeof": func(i interface{}) string {
		if i == nil {
			return "<nil>"
		}
		return reflect.TypeOf(i).String()
	},
	"inspect": func(i interface{}) string {
		if i == nil {
			return "<nil>"
		}
		return fmt.Sprintf("%+v", i)
	},
	"command": func(cb cmdCallback) goop.Command {
		return &cmd{cb}
	},
}

var _events = map[string]interface{}{
	"RunStart": network.RunStart{},
	"RunStop":  network.RunStop{},
	"Error":    &network.AsyncError{},

	"Channel":       &gateway.Channel{},
	"ConfigUpdate":  &gateway.ConfigUpdate{},
	"Connected":     &gateway.Connected{},
	"Disconnected":  &gateway.Disconnected{},
	"Clear":         &gateway.Clear{},
	"SystemMessage": &gateway.SystemMessage{},
	"Chat":          &gateway.Chat{},
	"PrivateChat":   &gateway.PrivateChat{},
	"Say":           &gateway.Say{},
	"Join":          &gateway.Join{},
	"UserUpdate":    &gateway.User{},
	"Leave":         &gateway.Leave{},
	"Trigger":       &gateway.Trigger{},

	"CapiPacket":          &capi.Packet{},
	"CapiAuthenticate":    &capi.Authenticate{},
	"CapiConnect":         &capi.Connect{},
	"CapiDisconnect":      &capi.Disconnect{},
	"CapiSendMessage":     &capi.SendMessage{},
	"CapiSendEmote":       &capi.SendEmote{},
	"CapiSendWhisper":     &capi.SendWhisper{},
	"CapiKickUser":        &capi.KickUser{},
	"CapiBanUser":         &capi.BanUser{},
	"CapiUnbanUser":       &capi.UnbanUser{},
	"CapiSetModerator":    &capi.SetModerator{},
	"CapiConnectEvent":    &capi.ConnectEvent{},
	"CapiDisconnectEvent": &capi.DisconnectEvent{},
	"CapiMessageEvent":    &capi.MessageEvent{},
	"CapiUserUpdateEvent": &capi.UserUpdateEvent{},
	"CapiUserLeaveEvent":  &capi.UserLeaveEvent{},

	"BncsKeepAlive":                      &bncs.KeepAlive{},
	"BncsPing":                           &bncs.Ping{},
	"BncsEnterChatReq":                   &bncs.EnterChatReq{},
	"BncsEnterChatResp":                  &bncs.EnterChatResp{},
	"BncsJoinChannel":                    &bncs.JoinChannel{},
	"BncsChatCommand":                    &bncs.ChatCommand{},
	"BncsChatEvent":                      &bncs.ChatEvent{},
	"BncsFloodDetected":                  &bncs.FloodDetected{},
	"BncsMessageBox":                     &bncs.MessageBox{},
	"BncsGetAdvListResp":                 &bncs.GetAdvListResp{},
	"BncsGetAdvListReq":                  &bncs.GetAdvListReq{},
	"BncsStartAdvex3Resp":                &bncs.StartAdvex3Resp{},
	"BncsStartAdvex3Req":                 &bncs.StartAdvex3Req{},
	"BncsStopAdv":                        &bncs.StopAdv{},
	"BncsNotifyJoin":                     &bncs.NotifyJoin{},
	"BncsNetGamePort":                    &bncs.NetGamePort{},
	"BncsAuthInfoResp":                   &bncs.AuthInfoResp{},
	"BncsAuthInfoReq":                    &bncs.AuthInfoReq{},
	"BncsAuthCheckResp":                  &bncs.AuthCheckResp{},
	"BncsAuthCheckReq":                   &bncs.AuthCheckReq{},
	"BncsAuthAccountCreateResp":          &bncs.AuthAccountCreateResp{},
	"BncsAuthAccountCreateReq":           &bncs.AuthAccountCreateReq{},
	"BncsAuthAccountLogonResp":           &bncs.AuthAccountLogonResp{},
	"BncsAuthAccountLogonReq":            &bncs.AuthAccountLogonReq{},
	"BncsAuthAccountLogonProofResp":      &bncs.AuthAccountLogonProofResp{},
	"BncsAuthAccountLogonProofReq":       &bncs.AuthAccountLogonProofReq{},
	"BncsAuthAccountChangePassResp":      &bncs.AuthAccountChangePassResp{},
	"BncsAuthAccountChangePassReq":       &bncs.AuthAccountChangePassReq{},
	"BncsAuthAccountChangePassProofResp": &bncs.AuthAccountChangePassProofResp{},
	"BncsAuthAccountChangePassProofReq":  &bncs.AuthAccountChangePassProofReq{},
	"BncsSetEmail":                       &bncs.SetEmail{},
	"BncsClanInfo":                       &bncs.ClanInfo{},
	"BncsUnknownPacket":                  &bncs.UnknownPacket{},

	"DiscordEvent":                    &discordgo.Event{},
	"DiscordConnect":                  &discordgo.Connect{},
	"DiscordDisconnect":               &discordgo.Disconnect{},
	"DiscordRateLimit":                &discordgo.RateLimit{},
	"DiscordReady":                    &discordgo.Ready{},
	"DiscordChannelCreate":            &discordgo.ChannelCreate{},
	"DiscordChannelUpdate":            &discordgo.ChannelUpdate{},
	"DiscordChannelDelete":            &discordgo.ChannelDelete{},
	"DiscordChannelPinsUpdate":        &discordgo.ChannelPinsUpdate{},
	"DiscordGuildCreate":              &discordgo.GuildCreate{},
	"DiscordGuildUpdate":              &discordgo.GuildUpdate{},
	"DiscordGuildDelete":              &discordgo.GuildDelete{},
	"DiscordGuildBanAdd":              &discordgo.GuildBanAdd{},
	"DiscordGuildBanRemove":           &discordgo.GuildBanRemove{},
	"DiscordGuildMemberAdd":           &discordgo.GuildMemberAdd{},
	"DiscordGuildMemberUpdate":        &discordgo.GuildMemberUpdate{},
	"DiscordGuildMemberRemove":        &discordgo.GuildMemberRemove{},
	"DiscordGuildRoleCreate":          &discordgo.GuildRoleCreate{},
	"DiscordGuildRoleUpdate":          &discordgo.GuildRoleUpdate{},
	"DiscordGuildRoleDelete":          &discordgo.GuildRoleDelete{},
	"DiscordGuildEmojisUpdate":        &discordgo.GuildEmojisUpdate{},
	"DiscordGuildMembersChunk":        &discordgo.GuildMembersChunk{},
	"DiscordGuildIntegrationsUpdate":  &discordgo.GuildIntegrationsUpdate{},
	"DiscordMessageAck":               &discordgo.MessageAck{},
	"DiscordMessageCreate":            &discordgo.MessageCreate{},
	"DiscordMessageUpdate":            &discordgo.MessageUpdate{},
	"DiscordMessageDelete":            &discordgo.MessageDelete{},
	"DiscordMessageReactionAdd":       &discordgo.MessageReactionAdd{},
	"DiscordMessageReactionRemove":    &discordgo.MessageReactionRemove{},
	"DiscordMessageReactionRemoveAll": &discordgo.MessageReactionRemoveAll{},
	"DiscordPresencesReplace":         &discordgo.PresencesReplace{},
	"DiscordPresenceUpdate":           &discordgo.PresenceUpdate{},
	"DiscordResumed":                  &discordgo.Resumed{},
	"DiscordRelationshipAdd":          &discordgo.RelationshipAdd{},
	"DiscordRelationshipRemove":       &discordgo.RelationshipRemove{},
	"DiscordTypingStart":              &discordgo.TypingStart{},
	"DiscordUserUpdate":               &discordgo.UserUpdate{},
	"DiscordUserSettingsUpdate":       &discordgo.UserSettingsUpdate{},
	"DiscordUserGuildSettingsUpdate":  &discordgo.UserGuildSettingsUpdate{},
	"DiscordUserNoteUpdate":           &discordgo.UserNoteUpdate{},
	"DiscordVoiceServerUpdate":        &discordgo.VoiceServerUpdate{},
	"DiscordVoiceStateUpdate":         &discordgo.VoiceStateUpdate{},
	"DiscordMessageDeleteBulk":        &discordgo.MessageDeleteBulk{},
}

var _errors = map[string]interface{}{
	"New": errors.New,
}

var _fmt = map[string]interface{}{
	"Errorf":   fmt.Errorf,
	"Fprint":   fmt.Fprint,
	"Fprintf":  fmt.Fprintf,
	"Fprintln": fmt.Fprintln,
	"Fscan":    fmt.Fscan,
	"Fscanf":   fmt.Fscanf,
	"Fscanln":  fmt.Fscanln,
	"Print":    fmt.Print,
	"Printf":   fmt.Printf,
	"Println":  fmt.Println,
	"Scan":     fmt.Scan,
	"Scanf":    fmt.Scanf,
	"Scanln":   fmt.Scanln,
	"Sprint":   fmt.Sprint,
	"Sprintf":  fmt.Sprintf,
	"Sprintln": fmt.Sprintln,
	"Sscan":    fmt.Sscan,
	"Sscanf":   fmt.Sscanf,
	"Sscanln":  fmt.Sscanln,
}

var _io = map[string]interface{}{
	"Copy":        io.Copy,
	"CopyBuffer":  io.CopyBuffer,
	"CopyN":       io.CopyN,
	"Pipe":        io.Pipe,
	"ReadAtLeast": io.ReadAtLeast,
	"ReadFull":    io.ReadFull,
	"WriteString": io.WriteString,
	"NopCloser":   ioutil.NopCloser,
	"ReadAll":     ioutil.ReadAll,
}
var _context = map[string]interface{}{
	"WithCancel":   context.WithCancel,
	"WithDeadline": context.WithDeadline,
	"WithTimeout":  context.WithTimeout,
	"Background":   context.Background,
	"TODO":         context.TODO,
}

var _time = map[string]interface{}{
	"ParseDuration": time.ParseDuration,
	"Since":         time.Since,
	"Until":         time.Until,
	"Date":          time.Date,
	"Now":           time.Now,
	"Parse":         time.Parse,
	"Unix":          time.Unix,

	"Nanosecond":  time.Nanosecond,
	"Microsecond": time.Microsecond,
	"Millisecond": time.Millisecond,
	"Second":      time.Second,
	"Minute":      time.Minute,
	"Hour":        time.Hour,
	"ANSIC":       time.ANSIC,
	"UnixDate":    time.UnixDate,
	"RubyDate":    time.RubyDate,
	"RFC822":      time.RFC822,
	"RFC822Z":     time.RFC822Z,
	"RFC850":      time.RFC850,
	"RFC1123":     time.RFC1123,
	"RFC1123Z":    time.RFC1123Z,
	"RFC3339":     time.RFC3339,
	"RFC3339Nano": time.RFC3339Nano,
	"Kitchen":     time.Kitchen,
	"Stamp":       time.Stamp,
	"StampMilli":  time.StampMilli,
	"StampMicro":  time.StampMicro,
	"StampNano":   time.StampNano,
}

var _bytes = map[string]interface{}{
	"Compare":         bytes.Compare,
	"Contains":        bytes.Contains,
	"ContainsAny":     bytes.ContainsAny,
	"ContainsRune":    bytes.ContainsRune,
	"Count":           bytes.Count,
	"Equal":           bytes.Equal,
	"EqualFold":       bytes.EqualFold,
	"Fields":          bytes.Fields,
	"FieldsFunc":      bytes.FieldsFunc,
	"HasPrefix":       bytes.HasPrefix,
	"HasSuffix":       bytes.HasSuffix,
	"Index":           bytes.Index,
	"IndexAny":        bytes.IndexAny,
	"IndexByte":       bytes.IndexByte,
	"IndexFunc":       bytes.IndexFunc,
	"IndexRune":       bytes.IndexRune,
	"Join":            bytes.Join,
	"LastIndex":       bytes.LastIndex,
	"LastIndexAny":    bytes.LastIndexAny,
	"LastIndexByte":   bytes.LastIndexByte,
	"LastIndexFunc":   bytes.LastIndexFunc,
	"Map":             bytes.Map,
	"Repeat":          bytes.Repeat,
	"Replace":         bytes.Replace,
	"Runes":           bytes.Runes,
	"Split":           bytes.Split,
	"SplitAfter":      bytes.SplitAfter,
	"SplitAfterN":     bytes.SplitAfterN,
	"SplitN":          bytes.SplitN,
	"Title":           bytes.Title,
	"ToLower":         bytes.ToLower,
	"ToTitle":         bytes.ToTitle,
	"ToUpper":         bytes.ToUpper,
	"Trim":            bytes.Trim,
	"TrimFunc":        bytes.TrimFunc,
	"TrimLeft":        bytes.TrimLeft,
	"TrimLeftFunc":    bytes.TrimLeftFunc,
	"TrimPrefix":      bytes.TrimPrefix,
	"TrimRight":       bytes.TrimRight,
	"TrimRightFunc":   bytes.TrimRightFunc,
	"TrimSpace":       bytes.TrimSpace,
	"TrimSuffix":      bytes.TrimSuffix,
	"NewBuffer":       bytes.NewBuffer,
	"NewBufferString": bytes.NewBufferString,
	"NewReader":       bytes.NewReader,
}

var _strings = map[string]interface{}{
	"Compare":       strings.Compare,
	"Contains":      strings.Contains,
	"ContainsAny":   strings.ContainsAny,
	"ContainsRune":  strings.ContainsRune,
	"Count":         strings.Count,
	"EqualFold":     strings.EqualFold,
	"Fields":        strings.Fields,
	"FieldsFunc":    strings.FieldsFunc,
	"HasPrefix":     strings.HasPrefix,
	"HasSuffix":     strings.HasSuffix,
	"Index":         strings.Index,
	"IndexAny":      strings.IndexAny,
	"IndexByte":     strings.IndexByte,
	"IndexFunc":     strings.IndexFunc,
	"IndexRune":     strings.IndexRune,
	"Join":          strings.Join,
	"LastIndex":     strings.LastIndex,
	"LastIndexAny":  strings.LastIndexAny,
	"LastIndexByte": strings.LastIndexByte,
	"LastIndexFunc": strings.LastIndexFunc,
	"Map":           strings.Map,
	"Repeat":        strings.Repeat,
	"Replace":       strings.Replace,
	"Split":         strings.Split,
	"SplitAfter":    strings.SplitAfter,
	"SplitAfterN":   strings.SplitAfterN,
	"SplitN":        strings.SplitN,
	"Title":         strings.Title,
	"ToLower":       strings.ToLower,
	"ToTitle":       strings.ToTitle,
	"ToUpper":       strings.ToUpper,
	"Trim":          strings.Trim,
	"TrimFunc":      strings.TrimFunc,
	"TrimLeft":      strings.TrimLeft,
	"TrimLeftFunc":  strings.TrimLeftFunc,
	"TrimPrefix":    strings.TrimPrefix,
	"TrimRight":     strings.TrimRight,
	"TrimRightFunc": strings.TrimRightFunc,
	"TrimSpace":     strings.TrimSpace,
	"TrimSuffix":    strings.TrimSuffix,
	"NewReader":     strings.NewReader,
	"NewReplacer":   strings.NewReplacer,
}

var _strconv = map[string]interface{}{
	"AppendBool":  strconv.AppendBool,
	"AppendFloat": strconv.AppendFloat,
	"AppendInt":   strconv.AppendInt,
	"AppendUint":  strconv.AppendUint,
	"Atoi":        strconv.Atoi,
	"FormatBool":  strconv.FormatBool,
	"FormatFloat": strconv.FormatFloat,
	"FormatInt":   strconv.FormatInt,
	"FormatUint":  strconv.FormatUint,
	"Itoa":        strconv.Itoa,
	"ParseBool":   strconv.ParseBool,
	"ParseFloat":  strconv.ParseFloat,
	"ParseInt":    strconv.ParseInt,
	"ParseUint":   strconv.ParseUint,
}

var _regex = map[string]interface{}{
	"Match":        regexp.Match,
	"MatchString":  regexp.MatchString,
	"QuoteMeta":    regexp.QuoteMeta,
	"Compile":      regexp.Compile,
	"CompilePOSIX": regexp.CompilePOSIX,
}

var _json = map[string]interface{}{
	"Marshal":       json.Marshal,
	"MarshalIndent": json.MarshalIndent,
	"Unmarshal":     json.Unmarshal,
}

var _sort = map[string]interface{}{
	"Float64s":          sort.Float64s,
	"Float64sAreSorted": sort.Float64sAreSorted,
	"Ints":              sort.Ints,
	"IntsAreSorted":     sort.IntsAreSorted,
	"IsSorted":          sort.IsSorted,
	"Search":            sort.Search,
	"SearchFloat64s":    sort.SearchFloat64s,
	"SearchInts":        sort.SearchInts,
	"SearchStrings":     sort.SearchStrings,
	"Slice":             sort.Slice,
	"SliceIsSorted":     sort.SliceIsSorted,
	"Strings":           sort.Strings,
}

var _url = map[string]interface{}{
	"PathEscape":      url.PathEscape,
	"PathUnescape":    url.PathUnescape,
	"QueryEscape":     url.QueryEscape,
	"QueryUnescape":   url.QueryUnescape,
	"Parse":           url.Parse,
	"ParseRequestURI": url.ParseRequestURI,
	"User":            url.User,
	"UserPassword":    url.UserPassword,
	"ParseQuery":      url.ParseQuery,
}

var _http = map[string]interface{}{
	"Get":      http.Get,
	"Head":     http.Head,
	"Post":     http.Post,
	"PostForm": http.PostForm,

	"StatusOK":                  http.StatusOK,
	"StatusBadRequest":          http.StatusBadRequest,
	"StatusUnauthorized":        http.StatusUnauthorized,
	"StatusForbidden":           http.StatusForbidden,
	"StatusNotFound":            http.StatusNotFound,
	"StatusRequestTimeout":      http.StatusRequestTimeout,
	"StatusInternalServerError": http.StatusInternalServerError,
	"StatusNotImplemented":      http.StatusNotImplemented,
	"StatusBadGateway":          http.StatusBadGateway,
	"StatusServiceUnavailable":  http.StatusServiceUnavailable,
	"StatusGatewayTimeout":      http.StatusGatewayTimeout,
}

var _websocket = map[string]interface{}{
	"Dialer": func() *websocket.Dialer { var d = *websocket.DefaultDialer; return &d },

	"TextMessage":   websocket.TextMessage,
	"BinaryMessage": websocket.BinaryMessage,
	"CloseMessage":  websocket.CloseMessage,
	"PingMessage":   websocket.PingMessage,
	"PongMessage":   websocket.PongMessage,
}

var _reflect = map[string]interface{}{
	"Copy":            reflect.Copy,
	"DeepEqual":       reflect.DeepEqual,
	"ArrayOf":         reflect.ArrayOf,
	"ChanOf":          reflect.ChanOf,
	"FuncOf":          reflect.FuncOf,
	"MapOf":           reflect.MapOf,
	"PtrTo":           reflect.PtrTo,
	"SliceOf":         reflect.SliceOf,
	"StructOf":        reflect.StructOf,
	"TypeOf":          reflect.TypeOf,
	"Append":          reflect.Append,
	"AppendSlice":     reflect.AppendSlice,
	"Indirect":        reflect.Indirect,
	"MakeChan":        reflect.MakeChan,
	"MakeFunc":        reflect.MakeFunc,
	"MakeMap":         reflect.MakeMap,
	"MakeMapWithSize": reflect.MakeMapWithSize,
	"MakeSlice":       reflect.MakeSlice,
	"New":             reflect.New,
	"Select":          reflect.Select,
	"ValueOf":         reflect.ValueOf,
	"Zero":            reflect.Zero,

	"Invalid":       reflect.Invalid,
	"Bool":          reflect.Bool,
	"Int":           reflect.Int,
	"Int8":          reflect.Int8,
	"Int16":         reflect.Int16,
	"Int32":         reflect.Int32,
	"Int64":         reflect.Int64,
	"Uint":          reflect.Uint,
	"Uint8":         reflect.Uint8,
	"Uint16":        reflect.Uint16,
	"Uint32":        reflect.Uint32,
	"Uint64":        reflect.Uint64,
	"Uintptr":       reflect.Uintptr,
	"Float32":       reflect.Float32,
	"Float64":       reflect.Float64,
	"Complex64":     reflect.Complex64,
	"Complex128":    reflect.Complex128,
	"Array":         reflect.Array,
	"Chan":          reflect.Chan,
	"Func":          reflect.Func,
	"Interface":     reflect.Interface,
	"Map":           reflect.Map,
	"Ptr":           reflect.Ptr,
	"Slice":         reflect.Slice,
	"String":        reflect.String,
	"Struct":        reflect.Struct,
	"UnsafePointer": reflect.UnsafePointer,
}
