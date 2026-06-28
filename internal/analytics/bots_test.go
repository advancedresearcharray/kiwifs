package analytics

import "testing"

func TestIsBot(t *testing.T) {
	tests := []struct {
		ua   string
		want bool
	}{
		// Known bots
		{"Googlebot/2.1 (+http://www.google.com/bot.html)", true},
		{"Mozilla/5.0 (compatible; bingbot/2.0)", true},
		{"Baiduspider/2.0", true},
		{"YandexBot/3.0", true},
		{"DuckDuckBot/1.0", true},
		{"facebookexternalhit/1.1", true},
		{"Twitterbot/1.0", true},
		{"LinkedInBot/1.0", true},
		{"WhatsApp/2.23", true},
		{"Applebot/0.1", true},
		{"curl/7.68.0", true},
		{"wget/1.20.3", true},
		{"python-requests/2.25.1", true},
		{"Go-http-client/1.1", true},
		{"Apache-HttpClient/4.5.13", true},
		{"Java/11.0.11", true},
		{"okhttp/4.9.0", true},
		{"Scrapy/2.5.0", true},
		{"HeadlessChrome/96.0.4664.110", true},
		{"PhantomJS/2.1.1", true},
		{"Selenium/4.0.0", true},
		{"SemrushBot/7", true},
		{"AhrefsBot/7.0", true},
		{"UptimeRobot/2.0", true},
		// Generic patterns
		{"some-random-bot", true},
		{"my-crawler/1.0", true},
		{"fast-spider", true},

		// Normal browsers
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36", false},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0", false},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 Safari/604.1", false},

		// Empty UA
		{"", false},
	}

	for _, tt := range tests {
		got := IsBot(tt.ua)
		if got != tt.want {
			t.Errorf("IsBot(%q) = %v, want %v", tt.ua, got, tt.want)
		}
	}
}

func BenchmarkIsBot(b *testing.B) {
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36"
	for i := 0; i < b.N; i++ {
		IsBot(ua)
	}
}

func BenchmarkIsBot_Bot(b *testing.B) {
	ua := "Googlebot/2.1 (+http://www.google.com/bot.html)"
	for i := 0; i < b.N; i++ {
		IsBot(ua)
	}
}
