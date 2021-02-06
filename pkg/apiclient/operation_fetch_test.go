package apiclient_test

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/fkondej/go-showcase/v1/pkg/apiclient"
	"github.com/fkondej/go-showcase/v1/pkg/libtest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("The AccountClient", func() {
	var (
		server              *ghttp.Server
		fetchPath           string
		accountClientConfig apiclient.AccountClientConfig
		accountClient       *apiclient.AccountClient
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		fetchPath = "/v1/accounts"
		serverURL, err := url.Parse(server.URL())
		Ω(err).ShouldNot(HaveOccurred())
		serverURL.Path = fetchPath
		accountClientConfig = apiclient.AccountClientConfig{
			URL:      serverURL.String(),
			ProxyURL: nil,
			Timeout:  time.Second,
		}
		accountClient = apiclient.NewAccountClient(&accountClientConfig)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Fetch Account Information operation", func() {
		type FetchAccountResponse struct {
			Data apiclient.AccountResource `json:"data"`
		}
		type FetchErrorResponse struct {
			ErrorMessage string `json:"error_message"`
		}

		var (
			accountID          string
			responseData       interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			accountID = libtest.GenerateID()
			responseData = nil
			responseStatusCode = http.StatusInternalServerError
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", path.Join(fetchPath, accountID)),
					ghttp.RespondWithJSONEncodedPtr(&responseStatusCode, &responseData),
				),
			)
		})

		Context("when Account exists", func() {

			BeforeEach(func() {
				responseStatusCode = http.StatusOK // 200
				responseData = &FetchAccountResponse{
					Data: apiclient.AccountResource{
						Type:           "accounts",
						ID:             accountID,
						OrganisationID: "eb0bd6f5-c3f5-44b2-b677-acd23cdde73c",
						Version:        0,
						Attributes:     libtest.GenerateAccountAttributes(),
					},
				}
			})

			It("should return Account Information without error", func() {
				accountData, err := accountClient.Fetch(accountID)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(accountData).ShouldNot(BeNil())
				Ω(server.ReceivedRequests()).Should(HaveLen(1))

				Ω(*accountData).Should(Equal(responseData.(*FetchAccountResponse).Data))
				Ω(accountData.Attributes.BIC).ShouldNot(BeNil())
				Ω(*accountData.Attributes.BaseCurrency).Should(HaveLen(3))
			})
		})

		Context("when Account does not exist", func() {

			BeforeEach(func() {
				responseStatusCode = http.StatusNotFound // 404
				responseData = FetchErrorResponse{
					ErrorMessage: fmt.Sprintf("record %v does not exist", accountID),
				}
			})

			It("should return nil result and ErrNoAccount error", func() {
				accountData, err := accountClient.Fetch(accountID)
				Ω(err).Should(MatchError(apiclient.ErrNoAccount))
				Ω(accountData).Should(BeNil())
				Ω(server.ReceivedRequests()).Should(HaveLen(1))
			})
		})
	})
})
