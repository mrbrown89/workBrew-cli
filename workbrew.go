package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const workbrewAPIVersion = "v0"

func workbrewGet(config Config, token string, endpoint string) (any, error) {
	url := fmt.Sprintf("%s/%s", config.URL, endpoint)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Workbrew-API-Version", workbrewAPIVersion)
	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Workbrew API returned %s", response.Status)
	}

	var result any

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
