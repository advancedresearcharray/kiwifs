package analytics

import "strings"

// knownBotPrefixes identifies common bot/crawler User-Agent prefixes.
// Requests from these agents are excluded from analytics to avoid
// inflating view and search counts.
var knownBotPrefixes = []string{
	"Googlebot",
	"bingbot",
	"Baiduspider",
	"YandexBot",
	"DuckDuckBot",
	"Slurp",
	"facebookexternalhit",
	"Twitterbot",
	"LinkedInBot",
	"WhatsApp",
	"Applebot",
	"MJ12bot",
	"AhrefsBot",
	"SemrushBot",
	"DotBot",
	"PetalBot",
	"UptimeRobot",
	"curl/",
	"wget/",
	"python-requests/",
	"Go-http-client/",
	"Apache-HttpClient/",
	"Java/",
	"okhttp/",
	"Scrapy/",
	"HeadlessChrome",
	"PhantomJS",
	"Selenium",
	"bot",
	"crawl",
	"spider",
}

// IsBot returns true if the User-Agent header looks like a known bot/crawler.
func IsBot(ua string) bool {
	if ua == "" {
		return false
	}
	lower := strings.ToLower(ua)
	for _, prefix := range knownBotPrefixes {
		if strings.Contains(lower, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}
