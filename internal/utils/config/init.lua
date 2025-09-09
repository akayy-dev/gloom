config = {
	tickers = {
		"SPY", "FEZ", "AAPL", "AMZN", "GOOGL", "MSFT", "NVDA", "META"
	},
	rss_feeds = {
		"https://www.nasdaq.com/feed/nasdaq-original/rss.xml",
		{
			url = "https://www.ft.com/myft/following/b29c92f0-ea01-4ad3-8ba3-0ca2fa488969.rss",
			cookie = "myFTLoginCookie"
		}
	},
	theme = {
		accent_color = "#703FFD",
	},
}

feeds = {
		"https://www.nasdaq.com/feed/nasdaq-original/rss.xml",
		{
			url = "https://www.ft.com/myft/following/b29c92f0-ea01-4ad3-8ba3-0ca2fa488969.rss",
			cookie = "myFTLoginCookie"
		}
}

gloom.accent_color = "#703FFD"
gloom.rss_feeds = feeds