package apiclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

type AccountClient struct {
	config     *AccountClientConfig
	httpClient *http.Client
}

// NewAccountClient creates a new instance of AccountClient
func NewAccountClient(config *AccountClientConfig) *AccountClient {
	transport := &http.Transport{
		Proxy: http.ProxyURL(config.ProxyURL),
	}

	httpClient := http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &AccountClient{
		config:     config,
		httpClient: &httpClient,
	}

}

//
// CONFIG
//

// AccountClientConfig is a configuration used to create `AccountClient`
type AccountClientConfig struct {
	URL      string        // Account API url address
	ProxyURL *url.URL      // Proxy to use when connecting to Account API
	Timeout  time.Duration // HTTP connection timeout
}

// getURL is a helper function that takes `AccountClientConfig.URL` and adds a subpath and query parameters
func (c *AccountClientConfig) getURL(subpath string, query url.Values) (string, error) {
	serverURL, err := url.Parse(c.URL)
	if err != nil {
		return "", fmt.Errorf("Wrong server URL %v: %v", c.URL, err)
	}
	endpointURL := url.URL{
		Scheme:   serverURL.Scheme,
		Host:     serverURL.Host,
		Path:     path.Join(serverURL.Path, subpath),
		RawQuery: query.Encode(),
	}
	return endpointURL.String(), nil
}

//
//  ERROR
//
var (
	// ErrConnection is returned when the Account Client fails to connect to the Server API
	// That includes: server unavailable and connection timeout
	ErrConnection = errors.New("Failed to connect to API server: please check your network connectivity")

	// ErrInternal is a general error, returned when the client encounters an unexpected problem
	ErrInternal = errors.New("API internal error: please report this to our support")

	// ErrNoAccount is returned when account does not exist or no accounts are available
	//  - `AccountClient.Fetch` returns the `ErrNoAccount` when there is no account with specified ID
	//  - `AccountClient.List` returns the `ErrNoAccount` for pages without a single account
	ErrNoAccount = errors.New("There is no account")

	// ErrAccountExist is returned when account with specified ID already exists
	//  - `AccountClient.Create` returns the `ErrAccountExist` when failed to create an account because it already exsits
	ErrAccountExist = errors.New("Account already exists")

	// ErrWrongVersion is returned when account exists but it has different version than specified
	//  - `AccountClient.Delete` returns the `ErrWrongVersion` when failed to delete an account because of version mismatch
	ErrWrongVersion = errors.New("Wrong Account version")

	// ErrWrongConfig is returned when `AccountClientConfig` contains not valid configuration
	ErrWrongConfig = errors.New("Wrong config")
)

//
//  LIST functionality
//

// AccountListFilter is a collection of filters used by `AccountClient.List` operation
type AccountListFilter struct {
	AccountNumber []string
	BankID        []string
	BankIDCode    []string
	Country       []string
	CustomerID    []string
	IBAN          []string
}

// AccountPage is an argument used for `AccountClient.List` operation
// It contains information used to request data in the first request
type AccountPage struct {
	PageNumber int               // Which page should be requested first, starts with 0
	PageSize   int               // Number of Accounts on each page
	Filter     AccountListFilter // Filters used to filter results
}

// FirstPage is a helper to request first page with default size
var FirstPage = AccountPage{PageNumber: 0, PageSize: 100}

// AccountPageResult is returned by `AccountClient.List` operation.
//
// There is no request send upon creation. First request is sent when `AccountPageResult.Next()` function is called.
//
// Usage:
//   accounts := client.List(page)
//   for accounts.Next() {
//     data, err := accounts.Data()
//     // if no err then `data` contains `[]AccountResource`
//   }
type AccountPageResult struct {
	currPage  AccountPage
	currData  []AccountResource
	lastError error
	loadData  func(page AccountPage) ([]AccountResource, error)
}

// Next send request to Account API and fetches data for the next page.
// It returns:
//  - `true` - if the fetch was successful and there is more data, or when the current fetch failed (when it happens to call Data() will return the error),
//  - `false` - if the fetch returned no more data, or there was an error with previous `Next()` calls.
//
// Next() remembers errors, and when it is called after the error occured then it will return `false`.
func (c *AccountPageResult) Next() bool {
	if c.lastError != nil {
		return false
	}
	if c.currData != nil {
		c.currPage.PageNumber++
		c.currData = nil
	}
	nextData, err := c.loadData(c.currPage)
	if err != nil {
		c.lastError = err
		if errors.Is(err, ErrNoAccount) {
			return false
		}
		return true
	}
	c.currData = nextData
	return true
}

// Data returns Account Information list fetched with the last `Next()` call.
// If last `Next()` call failed, then `Data()` returns error specyfing why `Next()` failed.
//
// `Data()` will aslo retturn error if called before first `Next()`, or after `Next()` returned `false`.
//
// Returns errors:
//   - ErrWrongConfig when Server URL is malformed
//   - ErrConnection when there are problems connecting Accounts API, e.g. server unavailable, or connection timeout
//   - ErrNoAccount when there is no more accounts, after `Next()` returned `false`
//   - ErrInternal when other, not handled issues appear
func (c *AccountPageResult) Data() ([]AccountResource, error) {
	if c.lastError != nil {
		return nil, c.lastError
	}
	if c.currData == nil {
		return nil, fmt.Errorf("Data not requested yet for the first time")
	}
	return c.currData, nil
}

// FetchAll fetch all accounts from Account API and put them into channel
//
// There is only one request to the Accounts API at the time. When one is done then next starts immediately
func (c *AccountPageResult) FetchAll() <-chan AccountResource {
	// queue up to three responses
	respCh := make(chan []AccountResource, 3)
	go func() {
		defer close(respCh)
		for c.Next() {
			data, err := c.Data()
			if err != nil {
				return
			}
			respCh <- data
		}
	}()

	accCh := make(chan AccountResource)
	go func() {
		defer close(accCh)
		for data := range respCh {
			for _, acc := range data {
				accCh <- acc
			}
		}
	}()

	return accCh
}
