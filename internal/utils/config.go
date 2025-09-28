package utils

import (
	"embed"
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type RSSFeed struct {
	URL    string
	Cookie string
}

type ThemeSettings struct {
	AccentColor      string
	PositiveMovement string
	NegativeMovement string
}

type UserConfig struct {
	APIKeys          map[string]string
	WatchlistTickers []string
	AccentColor      string
	Theme            ThemeSettings
	Tickers          []string
	RSSFeeds         []RSSFeed
}

var (
	// Represents the Users Config.
	Config UserConfig
	// The state of the Lua VM.
	LuaState = lua.NewState()
)

//go:embed config/*
var luaFS embed.FS

func RegisterLuaBindings() {
	LuaState.SetGlobal("notify", LuaState.NewFunction(func(L *lua.LState) int {
		prompt := L.CheckString(1)
		displayTime := L.CheckInt(2)

		Notify(prompt, displayTime)
		return 0
	}))
}

// Add a given path to the lua package path to allow requiring files from there.
func AddLuaPackagePath(dirPath string, configPath string) {
	packagePath := fmt.Sprintf(`
		package.path = package.path .. ";%s/?.lua;%s/?/init.lua"
		`, dirPath, configPath)
	LuaState.DoString(packagePath)
}

func LoadUserLuaConfig(path string) {
	RegisterLuaBindings()

	// Set the package path to the config directory.
	pathParts := strings.Split(path, "/")
	dirPath := strings.Join(pathParts[:len(pathParts)-1], "/")
	AddLuaPackagePath(dirPath, path)

	// Gloom Global
	cfg := LuaState.GetGlobal("gloom").(*lua.LTable)

	err := LuaState.DoFile(path)
	if err != nil {
		UserLog.Fatalf("Error: could not load config file at %s, %v", path, err)
	}

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
	// Read the theme table and set the Theme attribute
	if themeVal := cfg.RawGetString("theme"); themeVal.Type() == lua.LTTable {
		themeTable := themeVal.(*lua.LTable)
		if accent := themeTable.RawGetString("accent_color"); accent.Type() == lua.LTString {
			Config.Theme.AccentColor = accent.String()
		}
		if up := themeTable.RawGetString("up"); up.Type() == lua.LTString {
			Config.Theme.PositiveMovement = up.String()
		}
		if down := themeTable.RawGetString("down"); down.Type() == lua.LTString {
			Config.Theme.NegativeMovement = down.String()
		}
		UserLog.Infof("Theme loaded: %+v", Config.Theme)
	}

	watchlistVal := cfg.RawGetString("watchlist")
	watchlist, ok := watchlistVal.(*lua.LTable) // Make sure the type is a lua table
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

	// Get the API Keys
	apiKeys := cfg.RawGetString("api_keys")
	apiKeyTable, ok := apiKeys.(*lua.LTable)
	if !ok {
		UserLog.Warn("No 'api_keys' table found in config")
		return
	}

	// Before adding to table, make sure the hashmap exists
	if Config.APIKeys == nil {
		Config.APIKeys = make(map[string]string)
	}

	apiKeyTable.ForEach(func(_, v lua.LValue) {
		if v.Type() == lua.LTTable {
			tbl := v.(*lua.LTable)
			name := tbl.RawGetString("name").String()
			apiKey := tbl.RawGetString("key").String()
			Config.APIKeys[name] = apiKey

			UserLog.Infof("Found key for %s : %s", name, apiKey)
		}
	})

}

func LoadDefaultLuaConfig() {
	cfgTable := LuaState.NewTable()
	LuaState.SetGlobal("gloom", cfgTable) // add the global

	// SECTION: Create a watchlist.tickers table
	watchlistTable := LuaState.NewTable()
	LuaState.SetField(cfgTable, "watchlist", watchlistTable)
	tickersTable := LuaState.NewTable()
	LuaState.SetField(watchlistTable, "tickers", tickersTable)

	// SECTION: Load RSS Feeds table
	rssFeedsTable := LuaState.NewTable()
	LuaState.SetField(cfgTable, "rss_feeds", rssFeedsTable)

	// SECTION: Load theme Table
	themeTable := LuaState.NewTable()
	LuaState.SetField(cfgTable, "theme", themeTable)

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
