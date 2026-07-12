package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type shodan struct{ apiKey string }

func NewShodan(apiKey string) Searcher {
	return &shodan{apiKey: apiKey}
}

func (s *shodan) Name() string { return "Shodan" }

func (s *shodan) Search(query string, page int) (*SearchResult, error) {
	u := fmt.Sprintf("https://api.shodan.io/shodan/host/search?key=%s&query=%s&page=%d",
		s.apiKey, url.QueryEscape(query), page)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var apiResp struct {
		Total   int `json:"total"`
		Matches []struct {
			IPStr    string   `json:"ip_str"`
			Port     int      `json:"port"`
			Protocol string   `json:"protocol"`
			Hostnames []string `json:"hostnames"`
		} `json:"matches"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	result := &SearchResult{Total: apiResp.Total}
	for _, m := range apiResp.Matches {
		result.Results = append(result.Results, ResultItem{
			IP: m.IPStr, Port: m.Port, Protocol: m.Protocol,
		})
	}
	return result, nil
}
