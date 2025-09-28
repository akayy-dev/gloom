gloom.api_keys = {
	-- Source the gemini and financial modeling prep key from environment variables.
	-- You can substitute this with a plain string.
	{
		name = "gemini",
		key = os.getenv("GEMINI_KEY")
	},
	{
		name = "fmp",
		key = os.getenv("FMP_KEY")
	}
}

gloom.rss_feeds = {
	-- List of RSS feeds to show in the news table.
	"https://www.nasdaq.com/feed/nasdaq-original/rss.xml",
}

-- This table configures the stock tickers that will be displayed on the 
gloom.watchlist = {
	tickers = "SPY", "FEZ", "AAPL", "AMZN", "GOOGL", "MSFT", "NVDA", "META"
}

-- Theme settings
gloom.theme.accent_color = "#703FFD"
