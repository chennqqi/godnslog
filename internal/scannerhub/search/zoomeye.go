package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type zoomEye struct{ apiKey string }

func NewZoomEye(apiKey string) Searcher {
	return &zoomEye{apiKey: apiKey}
}

func (z *zoomEye) Name() string { return "ZoomEye" }

func (z *zoomEye) Search(query string, page int) (*SearchResult, error) {
	u := fmt.Sprintf("https://api.zoomeye.org/host/search?query=%s&page=%d", url.QueryEscape(query), page)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("API-KEY", z.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var apiResp struct {
		Total   int `json:"total"`
		Matches []struct {
			IP       string `json:"ip"`
			PortInfo struct {
				Port int `json:"port"`
			} `json:"portinfo"`
		} `json:"matches"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	result := &SearchResult{Total: apiResp.Total}
	for _, m := range apiResp.Matches {
		result.Results = append(result.Results, ResultItem{IP: m.IP, Port: m.PortInfo.Port})
	}
	return result, nil
}
