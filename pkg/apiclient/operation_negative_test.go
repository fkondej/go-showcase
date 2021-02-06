package apiclient_test

import (
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/fkondej/go-showcase/v1/pkg/apiclient"
	"github.com/fkondej/go-showcase/v1/pkg/libtest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/onsi/gomega/types"
)

var _ = Describe("[Negative cases] The AccountClient", func() {
	var (
		// API server
		server *ghttp.Server
		// API clients
		fetchPath     string
		clientConfig  apiclient.AccountClientConfig
		accountClient *apiclient.AccountClient

		// helper account info
		accountID         string
		version           int
		organisationID    string
		accountAttributes *apiclient.AccountAttributes

		// API client calls
		fetchOperation = func() (interface{}, error) { return accountClient.Fetch(accountID) }
		listOperation  = func() (interface{}, error) {
			accounts := accountClient.List(apiclient.FirstPage)
			accounts.Next()
			data, err := accounts.Data()
			return data, err
		}
		createOperation = func() (interface{}, error) {
			return accountClient.Create(accountID, organisationID, accountAttributes)
		}
		deleteOperation = func() (interface{}, error) { return accountClient.Delete(accountID, version) }
	)

	BeforeEach(func() {
		// Setup Server
		server = ghttp.NewServer()
		// Setup API clients
		fetchPath = "/v1/accounts"
		serverURL, err := url.Parse(server.URL())
		Ω(err).ShouldNot(HaveOccurred())
		serverURL.Path = fetchPath
		clientConfig = apiclient.AccountClientConfig{
			URL:      serverURL.String(),
			ProxyURL: nil,
			Timeout:  time.Second,
		}
		accountClient = apiclient.NewAccountClient(&clientConfig)
		// Setup helper Account Info
		accountID = libtest.GenerateID()
		version = rand.Intn(999)
		organisationID = libtest.GenerateOrganisationID()
		accountAttributes = libtest.GenerateAccountAttributes()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Server API side problems", func() {

		Context("when Server API is too slow", func() {
			var (
				wg *sync.WaitGroup
			)

			BeforeEach(func() {
				// need to create new clients in order to modify its Timeout
				clientConfig.Timeout = time.Millisecond * 20
				accountClient = apiclient.NewAccountClient(&clientConfig)
				accountClient = apiclient.NewAccountClient(&clientConfig)

				// "sleep" handler until the end of test case
				wg = &sync.WaitGroup{}
				wg.Add(1)
				server.AppendHandlers(func(rw http.ResponseWriter, r *http.Request) {
					wg.Wait() // use sync to minimize wait time
				})
			})

			DescribeTable("should return ErrConnection error",
				func(operation func() (interface{}, error), expectedResult types.GomegaMatcher) {
					defer wg.Done() // wake up handler at the test case end
					result, err := operation()
					Ω(err).Should(MatchError(apiclient.ErrConnection))
					Ω(result).Should(expectedResult)
					Ω(server.ReceivedRequests()).Should(HaveLen(1))
				},
				Entry("[Fetch operation]", fetchOperation, BeNil()),
				Entry("[List operation]", listOperation, BeNil()),
				Entry("[Create operation]", createOperation, BeNil()),
				Entry("[Delete operation]", deleteOperation, BeFalse()),
			)
		})

		Context("when Server API is down", func() {

			BeforeEach(func() {
				server.Close()
			})

			DescribeTable("should return ErrConnection error",
				func(operation func() (interface{}, error), expectedResult types.GomegaMatcher) {
					result, err := operation()
					Ω(err).Should(MatchError(apiclient.ErrConnection))
					Ω(result).Should(expectedResult)
					Ω(server.ReceivedRequests()).Should(HaveLen(0))
				},
				Entry("[Fetch operation]", fetchOperation, BeNil()),
				Entry("[List operation]", listOperation, BeNil()),
				Entry("[Create operation]", createOperation, BeNil()),
				Entry("[Delete operation]", deleteOperation, BeFalse()),
			)
		})

		Context("when Server API returns unhandled/unexpected status codes", func() {
			var (
				responseStatusCode int
			)

			BeforeEach(func() {
				server.AppendHandlers(ghttp.RespondWithPtr(&responseStatusCode, nil))
			})

			DescribeTable("should return ErrInternal error",
				func(operation func() (interface{}, error), expectedResult types.GomegaMatcher, statusCode int) {
					responseStatusCode = statusCode
					deleted, err := operation()
					Ω(err).Should(MatchError(apiclient.ErrInternal))
					Ω(deleted).Should(expectedResult)
					Ω(server.ReceivedRequests()).Should(HaveLen(1))
				},
				Entry("[Fetch operation] 500", fetchOperation, BeNil(), http.StatusInternalServerError),
				Entry("[Fetch operation] 400", fetchOperation, BeNil(), http.StatusBadRequest),
				Entry("[List operation] 500", listOperation, BeNil(), http.StatusInternalServerError),
				Entry("[List operation] 400", listOperation, BeNil(), http.StatusBadRequest),
				Entry("[Create operation] 500", createOperation, BeNil(), http.StatusInternalServerError),
				Entry("[Create operation] 400", createOperation, BeNil(), http.StatusBadRequest),
				Entry("[Delete operation] 500", deleteOperation, BeFalse(), http.StatusInternalServerError),
				Entry("[Delete operation] 400", deleteOperation, BeFalse(), http.StatusBadRequest),
				Entry("[Delete operation] 200", deleteOperation, BeFalse(), http.StatusOK),
			)
		})

		Context("when Server API returns malformed json in body", func() {
			var (
				responseStatusCode int
			)

			BeforeEach(func() {
				responseBody := "{\"aa\": \"missing rest..."
				server.AppendHandlers(ghttp.RespondWithPtr(&responseStatusCode, &responseBody))
			})

			DescribeTable("should return ErrInternal error",
				func(operation func() (interface{}, error), expectedResult types.GomegaMatcher, statusCode int) {
					responseStatusCode = statusCode
					deleted, err := operation()
					Ω(err).Should(MatchError(apiclient.ErrInternal))
					Ω(deleted).Should(expectedResult)
					Ω(server.ReceivedRequests()).Should(HaveLen(1))
				},
				Entry("[Fetch operation]", fetchOperation, BeNil(), http.StatusOK),
				Entry("[List operation]", listOperation, BeNil(), http.StatusOK),
				Entry("[Create operation]", createOperation, BeNil(), http.StatusCreated),
			)
		})
	})

	Describe("Client side issues", func() {

		Context("when Configuration has malformed server URL", func() {

			BeforeEach(func() {
				// need to create new clients in order to modify its Config
				clientConfig = apiclient.AccountClientConfig{
					URL:      "http://localhost:8080aaa/cc", // malformed URL
					ProxyURL: nil,
					Timeout:  time.Second,
				}
				accountClient = apiclient.NewAccountClient(&clientConfig)
				accountClient = apiclient.NewAccountClient(&clientConfig)
			})

			DescribeTable("should return ErrWrongConfig error",
				func(operation func() (interface{}, error), expectedResult types.GomegaMatcher) {
					operationResult, err := operation()
					Ω(err).Should(MatchError(apiclient.ErrWrongConfig))
					Ω(operationResult).Should(expectedResult)
					Ω(server.ReceivedRequests()).Should(HaveLen(0)) // should have not been called
				},
				Entry("[Fetch operation]", fetchOperation, BeNil()),
				Entry("[List operation]", listOperation, BeNil()),
				Entry("[Create operation]", createOperation, BeNil()),
				Entry("[Delete operation]", deleteOperation, BeFalse()),
			)
		})
	})
})
