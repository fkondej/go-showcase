package apiclient_test

import (
	"fmt"
	"math/rand"
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
		deletePath          string
		accountClientConfig apiclient.AccountClientConfig
		accountClient       *apiclient.AccountClient
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		deletePath = "/v1/accounts"
		serverURL, err := url.Parse(server.URL())
		Ω(err).ShouldNot(HaveOccurred())
		serverURL.Path = deletePath
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

	Describe("Delete Account operation", func() {
		var (
			accountID          string
			version            int
			responseStatusCode int
		)

		BeforeEach(func() {
			accountID = libtest.GenerateID()
			version = rand.Intn(999)
			responseStatusCode = http.StatusInternalServerError
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", path.Join(deletePath, accountID), fmt.Sprintf("version=%v", version)),
					ghttp.RespondWithPtr(&responseStatusCode, nil),
				),
			)
		})

		Context("when Account exists", func() {

			BeforeEach(func() {
				responseStatusCode = http.StatusNoContent // 204
			})

			It("should return true (deleted) without error", func() {
				deleted, err := accountClient.Delete(accountID, version)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(deleted).Should(BeTrue())
				Ω(server.ReceivedRequests()).Should(HaveLen(1))
			})
		})

		Context("when Account does not exist", func() {

			BeforeEach(func() {
				responseStatusCode = http.StatusNotFound // 404
			})

			It("should return false (not deleted) without error", func() {
				deleted, err := accountClient.Delete(accountID, version)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(deleted).Should(BeFalse())
				Ω(server.ReceivedRequests()).Should(HaveLen(1))
			})
		})

		Context("when Account exists but with different version", func() {

			BeforeEach(func() {
				responseStatusCode = http.StatusConflict // 409
			})

			It("should return false (not deleted) with ErrWrongVersion error", func() {
				deleted, err := accountClient.Delete(accountID, version)
				Ω(err).Should(MatchError(apiclient.ErrWrongVersion))
				Ω(deleted).Should(BeFalse())
				Ω(server.ReceivedRequests()).Should(HaveLen(1))
			})
		})
	})

})
