package video

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PexelsVideoResponse maps the JSON from Pexels API
type PexelsVideoResponse struct {
	Page         int `json:"page"`
	PerPage      int `json:"per_page"`
	TotalResults int `json:"total_results"`
	Videos       []struct {
		ID       int    `json:"id"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		URL      string `json:"url"`
		Image    string `json:"image"`
		Duration int    `json:"duration"`
		User     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"user"`
		VideoFiles []struct {
			ID       int    `json:"id"`
			Quality  string `json:"quality"`
			FileType string `json:"file_type"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
			Link     string `json:"link"`
		} `json:"video_files"`
	} `json:"videos"`
}

// FetchStockVideos queries the Pexels Stock Video API
func FetchStockVideos(apiKey, query string) (*PexelsVideoResponse, error) {
	if query == "" {
		query = "nature" // default fallback
	}
	url := fmt.Sprintf("https://api.pexels.com/videos/search?query=%s&per_page=15", query)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pexels API returned status: %d", resp.StatusCode)
	}

	var pexelsData PexelsVideoResponse
	if err := json.NewDecoder(resp.Body).Decode(&pexelsData); err != nil {
		return nil, err
	}

	return &pexelsData, nil
}
