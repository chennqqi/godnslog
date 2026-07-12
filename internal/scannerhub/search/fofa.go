package search

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type fofa struct{ email, key string }

func NewFofa(email, key string) Searcher {
	return &fofa{email: email, key: key}
}

func (f *fofa) Name() string { return "Fofa" }

func (f *fofa) Search(query string, page int) (*SearchResult, error) {
	qbase64 := base64.StdEncoding.EncodeToString([]byte(query))
	u := fmt.Sprintf("https://fofa.info/api/v1/search/all?email=%s&key=%s&qbase64=%s&page=%d&size=20",
		url.QueryEscape(f.email), url.QueryEscape(f.key), url.QueryEscape(qbase64), page)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var apiResp struct {
		Error   bool          `json:"error"`
		Results []interface{} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if apiResp.Error {
		return nil, fmt.Errorf("fofa API error")
	}
	result := &SearchResult{}
	// Fofa returns: [ip, port, protocol, hostname, country, title, ...]
	for _, row := range apiResp.Results {
		if cols, ok := row.([]interface{}); ok && len(cols) >= 2 {
			item := ResultItem{}
			if ip, ok := cols[0].(string); ok {
				item.IP = ip
			}
			if port, ok := cols[1].(float64); ok {
				item.Port = int(port)
			}
			result.Results = append(result.Results, item)
		}
	}
	result.Total = len(result.Results)
	return result, nil
}
