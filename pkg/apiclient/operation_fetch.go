package apiclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Fetch operation requests Account Information from underlying Accounts API
//
// Returns errors:
//   - ErrWrongConfig when Server URL is malformed
//   - ErrConnection when there are problems connecting Accounts API, e.g. server unavailable, or connection timeout
//   - ErrInternal when other, not handled issues appear
func (client *AccountClient) Fetch(accountID string) (*AccountResource, error) {
	// get Server URL
	fetchURL, err := client.config.getURL(accountID, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch account info: wrong API url %v %w", err, ErrWrongConfig)
	}
	// SEND request
	resp, err := client.httpClient.Get(fetchURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch account info: response error %v %w", err, ErrConnection)
	}
	defer resp.Body.Close()
	// check Response Status Codes
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Failed to find account %v %w", accountID, ErrNoAccount)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to fetch account info: response status code %v %w", resp.StatusCode, ErrInternal)
	}
	// parse Response Body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch account info: response body error %v %w", err, ErrInternal)
	}
	var jsonResponse struct {
		Data AccountResource `json:"data"`
	}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch account info: json parse issue %v %w", err, ErrInternal)
	}
	// return parsed Response
	return &jsonResponse.Data, nil
}
