package apiclient

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Delete operation requests Account deletion in underlying Accounts API
//
// Returns `true` only when the Account is deleted
// Returns `false` without `error` only when the Account does not exists
// Returns `false` with `error` when problems occured
//
// Errors:
//   - ErrWrongConfig when Server URL is malformed
//   - ErrConnection when there are problems connecting Accounts API, e.g. server unavailable, or connection timeout
//   - ErrWrongVersion when Account exists, but Accounts API did not delete it, because requested version of the Account was different
//   - ErrInternal when other, not handled issues appear
func (client *AccountClient) Delete(accountID string, version int) (bool, error) {
	// get Server URL (with Account id and version)
	query := url.Values{}
	query.Add("version", strconv.Itoa(version))
	deleteURL, err := client.config.getURL(accountID, query)
	if err != nil {
		return false, fmt.Errorf("Failed to delete account: wrong API url %v %w", err, ErrWrongConfig)
	}
	// prepare Request
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return false, fmt.Errorf("Failed to delete account: unknow error %v %w", err, ErrInternal)
	}
	// SEND the Request
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("Failed to delete account: response error %v %w", err, ErrConnection)
	}
	defer resp.Body.Close()
	// check Response Status Codes
	if resp.StatusCode == http.StatusNoContent { // 204 - Resource has been successfully deleted
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound { // 404 - Specified resource does not exist
		return false, nil
	}
	if resp.StatusCode == http.StatusConflict { // 409 - Specified version incorrect
		return false, fmt.Errorf("Failed to delete account: wrong version %v of the account %v %w", version, accountID, ErrWrongVersion)
	}
	// different Status Code
	return false, fmt.Errorf("Failed to delete account: wrong response %v %w", resp.StatusCode, ErrInternal)
}
