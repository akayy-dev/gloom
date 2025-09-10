package utils

import (
	"bytes"
	"embed"

	lua "github.com/yuin/gopher-lua"
)

type RSSFeed struct {
	URL    string
	Cookie string
}

type UserConfig struct {
	WatchlistTickers []string
	AccentColor      string
	Tickers          []string
	RSSFeeds         []RSSFeed
}

var (
	// Config manager
	Config   UserConfig
	LuaState = lua.NewState()
)

//go:embed config/*
var luaFS embed.FS

func LoadUserLuaConfig(path string) {
	err := LuaState.DoFile(path)
	if err != nil {
		UserLog.Fatalf("Error: could not load config file at %s, %v", path, err)
	}

	cfg := LuaState.GetGlobal("gloom").(*lua.LTable)

	ParseLuaConfig(*cfg)
}

// Parses the Lua config table and sets the Config variable to values declared.
func ParseLuaConfig(cfg lua.LTable) {
	feeds := cfg.RawGetString("rss_feeds")

	if tbl, ok := feeds.(*lua.LTable); ok {
		Config.RSSFeeds = loadRSSFeedsFromTable(*tbl)
		UserLog.Info("Got RSS Feed")
		UserLog.Info(Config.RSSFeeds)
	}

	// Read the accent color
	if accentColor := cfg.RawGetString("accent_color"); accentColor.Type() == lua.LTString {
		Config.AccentColor = accentColor.String()
		UserLog.Infof("Accent color from lua script is %s", Config.AccentColor)
	}

	watchlistVal := cfg.RawGetString("watchlist")
	watchlist, ok := watchlistVal.(*lua.LTable)
	if !ok || watchlist == nil {
		UserLog.Warn("No 'watchlist' table found in config")
		return
	}

	tickersVal := watchlist.RawGetString("tickers")
	tickers, ok := tickersVal.(*lua.LTable)
	if !ok || tickers == nil {
		UserLog.Warn("No 'tickers' table found in watchlist")
		return
	}

	// Set the value of the tickers to nil (because the config currently has the default tickers)
	Config.Tickers = nil
	tickers.ForEach(func(_, symbol lua.LValue) {
		switch symbol.Type() {
		case lua.LTString:
			Config.Tickers = append(Config.Tickers, symbol.String())
		}
	})
}

func LoadDefaultLuaConfig() {
	cfgTable := LuaState.NewTable()
	LuaState.SetGlobal("gloom", cfgTable) // add the global

	script, err := luaFS.ReadFile("config/init.lua")
	// TODO: Change panic to a log.
	if err != nil {
		panic(err)
	}
	if err := LuaState.DoString(string(script)); err != nil {
		panic(err)
	}
	ParseLuaConfig(*cfgTable)
}

func loadRSSFeedsFromTable(feedTable lua.LTable) []RSSFeed {
	feeds := []RSSFeed{}
	feedTable.ForEach(func(_, v lua.LValue) {
		// Check the type, if string it's just a URL
		// If table it has a cookie
		feed := RSSFeed{}
		switch v.Type() {
		case lua.LTString:
			feed.URL = v.String()
			feeds = append(feeds, feed)
		case lua.LTTable:
			tbl := v.(*lua.LTable)
			UserLog.Debug("Found feed table")
			feed.URL = tbl.RawGetString("url").String()
			if cookie := tbl.RawGetString("cookie"); cookie.Type() == lua.LTString {
				feed.Cookie = cookie.String()
			}
			feeds = append(feeds, feed)
		}
		UserLog.Infof("Found RSS Feed: %s", feed)
	})

	return feeds
}

// Takes the bytes from a JSON array and removes their comment lines (lines starting with //)
func StripCommentsFromJSON(fileContent []byte) ([]byte, error) {
	lines := bytes.Split(fileContent, []byte("\n"))
	var filteredLines [][]byte

	for _, line := range lines {
		trimmedLine := bytes.TrimSpace(line)
		if !bytes.HasPrefix(trimmedLine, []byte("//")) {
			filteredLines = append(filteredLines, line)
		}
	}

	return bytes.Join(filteredLines, []byte("\n")), nil
}
