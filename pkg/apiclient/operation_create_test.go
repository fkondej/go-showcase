package apiclient_test

import (
	"net/http"
	"net/url"
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
		createPath          string
		accountClientConfig apiclient.AccountClientConfig
		accountClient       *apiclient.AccountClient
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		createPath = "/v1/accounts"
		serverURL, err := url.Parse(server.URL())
		Ω(err).ShouldNot(HaveOccurred())
		serverURL.Path = createPath
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

	Describe("Create Account operation", func() {
		type CreateAccountResponse struct {
			Data apiclient.AccountResource `json:"data"`
		}
		type CreateErrorResponse struct {
			ErrorMessage string `json:"error_message"`
		}

		var (
			accountID         string
			organisationID    string
			accountAttributes *apiclient.AccountAttributes

			responseData       interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			accountID = libtest.GenerateID()
			organisationID = libtest.GenerateOrganisationID()
			accountAttributes = libtest.GenerateAccountAttributes()
			responseData = nil
			responseStatusCode = http.StatusInternalServerError
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", createPath),
					ghttp.RespondWithJSONEncodedPtr(&responseStatusCode, &responseData),
				),
			)
		})

		Context("when Account does not exist", func() {

			BeforeEach(func() {
				responseStatusCode = http.StatusCreated // 201
				responseData = &CreateAccountResponse{
					Data: apiclient.AccountResource{
						Type:           "accounts",
						ID:             accountID,
						OrganisationID: organisationID,
						Version:        0,
						Attributes:     accountAttributes,
					},
				}
			})

			It("should create an Account and return Account Information without error", func() {
				accountData, err := accountClient.Create(accountID, organisationID, accountAttributes)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(accountData).ShouldNot(BeNil())
				Ω(accountData.ID).Should(Equal(accountID))
				Ω(accountData.OrganisationID).Should(Equal(organisationID))
				Ω(accountData.Attributes.Country).Should(Equal(accountAttributes.Country))
				Ω(server.ReceivedRequests()).Should(HaveLen(1))
			})
		})

		Context("when Account exists", func() {

			BeforeEach(func() {
				responseStatusCode = http.StatusConflict // 409
				responseData = &CreateErrorResponse{
					ErrorMessage: "Account cannot be created as it violates a duplicate constraint",
				}
			})

			It("should return nil with ErrAccountExist error", func() {
				accountData, err := accountClient.Create(accountID, organisationID, accountAttributes)
				Ω(err).Should(MatchError(apiclient.ErrAccountExist))
				Ω(accountData).Should(BeNil())
				Ω(server.ReceivedRequests()).Should(HaveLen(1))
			})
		})

		Context("[negative] when Server API returns success (Account successfully created) but the returned account information is different to the requested", func() {

			BeforeEach(func() {
				responseStatusCode = http.StatusCreated // 201
				responseData = &CreateAccountResponse{
					Data: apiclient.AccountResource{
						Type:           "accounts",
						ID:             libtest.GenerateID(),
						OrganisationID: libtest.GenerateOrganisationID(),
						Version:        0,
						Attributes:     libtest.GenerateAccountAttributes(),
					},
				}
			})

			It("should return nil with ErrInternal error", func() {
				accountData, err := accountClient.Create(accountID, organisationID, accountAttributes)
				Ω(err).Should(MatchError(apiclient.ErrInternal))
				Ω(accountData).Should(BeNil())
				Ω(server.ReceivedRequests()).Should(HaveLen(1))
			})
		})
	})
})
