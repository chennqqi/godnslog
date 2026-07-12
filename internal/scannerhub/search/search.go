package search

// ResultItem represents a single search result
type ResultItem struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Country  string `json:"country,omitempty"`
	Title    string `json:"title,omitempty"`
}

// SearchResult represents search engine results
type SearchResult struct {
	Total   int          `json:"total"`
	Results []ResultItem `json:"results"`
}

// Searcher is the interface for search engine adapters
type Searcher interface {
	Name() string
	Search(query string, page int) (*SearchResult, error)
}

// Config holds API credentials for all search engines
type Config struct {
	ZoomEyeKey string
	ShodanKey  string
	FofaEmail  string
	FofaKey    string
}
