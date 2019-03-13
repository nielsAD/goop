// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package plugin

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/bwmarrin/discordgo"
	luar "github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
	"github.com/nielsAD/gowarcraft3/protocol/capi"
)

// goop.Command wrapper for plugins
type cmdCallback func(t *gateway.Trigger, gw gateway.Gateway) error
type cmd struct{ cb cmdCallback }

func (c *cmd) CanExecute(t *gateway.Trigger) bool                                 { return true }
func (c *cmd) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error { return c.cb(t, gw) }

func luaTypeOf(i interface{}) string {
	if i == nil {
		return "<nil>"
	}
	return reflect.TypeOf(i).String()
}

func luaInspect(i interface{}) string {
	if i == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%+v", i)
}

func luaError(s string) error {
	return errors.New(s)
}

func luaCommand(cb cmdCallback) goop.Command {
	return &cmd{cb}
}

func importFunctions(ls *lua.LState) {
	ls.SetGlobal("goerror", luar.New(ls, luaError))
	ls.SetGlobal("gotypeof", luar.New(ls, luaTypeOf))
	ls.SetGlobal("inspect", luar.New(ls, luaInspect))
	ls.SetGlobal("command", luar.New(ls, luaCommand))
}

func importEvents(ls *lua.LState) {

	// Define local so nothing is shared between plugins
	var events = map[string]interface{}{
		"RunStart": network.RunStart{},
		"RunStop":  network.RunStop{},
		"Error":    &network.AsyncError{},

		"User":          &gateway.User{},
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

	var tab = ls.NewTable()
	for k, t := range events {
		var str = strings.TrimLeft(luaTypeOf(t), "*")
		ls.SetField(tab, k, luar.New(ls, t))
		ls.SetField(tab, str, luar.NewType(ls, t))
	}

	ls.SetGlobal("events", tab)
}

func importAccess(ls *lua.LState) {
	var access = ls.NewTable()
	for i, str := range gateway.ConStrings {
		if len(str) == 0 {
			str = "Default"
		}
		str = strings.Title(str)
		ls.SetField(access, str, lua.LNumber(gateway.ConLevels[i]))
		ls.SetTable(access, lua.LNumber(gateway.ConLevels[i]), lua.LString(str))
	}

	ls.SetGlobal("access", access)
}
