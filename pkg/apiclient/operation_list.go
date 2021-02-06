package apiclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// List operation requests a list of Accounts Info with the ability to filter and page
//
// Filter - only Accounts that meet filter criteria are returned. Combination of filters act as AND expressions
//
// Page - provide List of all Accounts in equal chunks/pages. Each page has `pageSize` Accounts, and only the last page might have less.
//
// It returns `AccountPageResult` that allows getting all Pagest (starting from requestsd in `List()`), and Account List for every page
//
// `List()` does not send any requests to underlying Accounts API. The first request is sent when `AccountPageResult.Next()` is called for the first time.
func (client *AccountClient) List(page AccountPage) AccountPageResult {
	return AccountPageResult{
		currPage:  page,
		currData:  nil,
		lastError: nil,
		loadData:  client.fetchAccountList,
	}
}

func (client *AccountClient) fetchAccountList(page AccountPage) ([]AccountResource, error) {
	listURL, err := client.config.getURL("", convertToQuery(page.PageNumber, page.PageSize, page.Filter))
	if err != nil {
		return nil, fmt.Errorf("Failed to get list: wrong API url %v %w", err, ErrWrongConfig)
	}
	resp, err := client.httpClient.Get(listURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to get list: response error %v %w", err, ErrConnection)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get list: response status code %v %w", resp.StatusCode, ErrInternal)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to get list: response body error %v %w", err, ErrInternal)
	}
	var jsonResponse struct {
		Data []AccountResource `json:"data"`
	}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("Failed to get List: json parse issue %v %w", err, ErrInternal)
	}
	if len(jsonResponse.Data) == 0 {
		return nil, fmt.Errorf("No more accounts %w", ErrNoAccount)
	}
	return jsonResponse.Data, nil
}

func convertToQuery(pageNumber int, pageSize int, filter AccountListFilter) url.Values {
	values := url.Values{}
	values.Add("page[number]", strconv.Itoa(pageNumber))
	if pageSize <= 0 {
		values.Add("page[size]", "100")
	} else {
		values.Add("page[size]", strconv.Itoa(pageSize))
	}
	if len(filter.AccountNumber) > 0 {
		values.Add("filter[account_number]", strings.Join(filter.AccountNumber, ","))
	}
	if len(filter.BankID) > 0 {
		values.Add("filter[bank_id]", strings.Join(filter.BankID, ","))
	}
	if len(filter.BankIDCode) > 0 {
		values.Add("filter[bank_id_code]", strings.Join(filter.BankIDCode, ","))
	}
	if len(filter.Country) > 0 {
		values.Add("filter[country]", strings.Join(filter.Country, ","))
	}
	if len(filter.CustomerID) > 0 {
		values.Add("filter[customer_id]", strings.Join(filter.CustomerID, ","))
	}
	if len(filter.IBAN) > 0 {
		values.Add("filter[iban]", strings.Join(filter.IBAN, ","))
	}
	return values
}
