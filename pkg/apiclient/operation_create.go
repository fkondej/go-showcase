package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Create operation requests a new Account creation in underlying Accounts API.
//
// If successful then returns Account Information returned by Accounts API.
//
// Returns errors:
//   - ErrWrongConfig when Server URL is malformed
//   - ErrConnection when there are problems connecting Accounts API, e.g. server unavailable, or connection timeout
//   - ErrAccountExist when account with requeted accountID already exists
//   - ErrInternal when other, not handled issues appear
func (client *AccountClient) Create(accountID string, organisationID string, accountAttributes *AccountAttributes) (*AccountResource, error) {
	// get Server URL
	createURL, err := client.config.getURL("", nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create an account: wrong API url %v %w", err, ErrWrongConfig)
	}
	// prepare Request Data
	data := CreateAccountResourceRequestData{}
	data.Data.Type = "accounts"
	data.Data.ID = accountID
	data.Data.OrganisationID = organisationID
	data.Data.Attributes = accountAttributes

	strData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to create an account: failed to create request data %v %w", err, ErrInternal)
	}
	// SEND request
	resp, err := client.httpClient.Post(createURL, "application/json", bytes.NewBuffer(strData))
	if err != nil {
		return nil, fmt.Errorf("Failed to create an account: response error %v %w", err, ErrConnection)
	}
	defer resp.Body.Close()
	// check Response Status Codes
	if resp.StatusCode == http.StatusConflict { // 409
		return nil, fmt.Errorf("Failed to create an account: account with specified id already exists, %w", ErrAccountExist)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Failed to create an account: response status code %v %w", resp.StatusCode, ErrInternal)
	}
	// parse Response Body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to create an account: response body error %v %w", err, ErrInternal)
	}
	var jsonResponse struct {
		Data AccountResource `json:"data"`
	}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("Failed to create an account: response json parse issue %v %w", err, ErrInternal)
	}
	// some sanity check
	if jsonResponse.Data.ID != accountID || jsonResponse.Data.OrganisationID != organisationID ||
		jsonResponse.Data.Attributes.Country != accountAttributes.Country {
		// TODO: should we call our error API to informa about issue?
		return nil, fmt.Errorf("Failed to create an account: response data contains different data then requested. %w", ErrInternal)
	}
	// return parsed Response
	return &jsonResponse.Data, nil
}
