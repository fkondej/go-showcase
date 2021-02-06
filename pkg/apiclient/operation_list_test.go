package apiclient_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/fkondej/go-showcase/v1/pkg/apiclient"
	"github.com/fkondej/go-showcase/v1/pkg/libtest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("The AccountClient", func() {
	var (
		server              *ghttp.Server
		listPath            string
		accountClientConfig apiclient.AccountClientConfig
		accountClient       *apiclient.AccountClient
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		listPath = "/v1/accounts"
		serverURL, err := url.Parse(server.URL())
		Ω(err).ShouldNot(HaveOccurred())
		serverURL.Path = listPath
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

	type ListAccountResponse struct {
		Data []apiclient.AccountResource `json:"data"`
	}

	Describe("List Accounts: request single page", func() {
		var (
			rawQuery           []string
			responseData       interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			rawQuery = []string{"broken"}
			responseData = nil
			responseStatusCode = http.StatusInternalServerError
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", listPath, rawQuery...),
					ghttp.RespondWithJSONEncodedPtr(&responseStatusCode, &responseData),
				),
			)
		})

		Describe("request First Page", func() {
			var (
				listOfAccounts []apiclient.AccountResource
			)

			Context("when accounts exist", func() {

				BeforeEach(func() {
					responseStatusCode = http.StatusOK // 200
					listOfAccounts = libtest.GenerateAccountResources(35)
					responseData = &ListAccountResponse{
						Data: listOfAccounts,
					}
					rawQuery[0] = "page[number]=0&page[size]=100"
				})

				DescribeTable("should return page with accounts without error",
					func(firstPage apiclient.AccountPage) {
						accounts := accountClient.List(apiclient.FirstPage)
						somethingToRead := accounts.Next()
						Ω(somethingToRead).Should(BeTrue())
						resultPageData, err := accounts.Data()
						Ω(err).ShouldNot(HaveOccurred())
						Ω(resultPageData).ShouldNot(BeNil())
						Ω(resultPageData).ShouldNot(BeNil())
						Ω(resultPageData).Should(HaveLen(len(listOfAccounts)))
						Ω(resultPageData).Should(Equal(listOfAccounts))

						Ω(server.ReceivedRequests()).Should(HaveLen(1))
					},
					Entry("FirstPage", apiclient.FirstPage),
					Entry("default page should behave as first page", apiclient.AccountPage{}),
				)
			})

			Context("when no account exists", func() {

				BeforeEach(func() {
					responseStatusCode = http.StatusOK // 200
					responseData = &ListAccountResponse{
						Data: []apiclient.AccountResource{},
					}
					rawQuery[0] = "page[number]=0&page[size]=100"
				})

				DescribeTable("should return nil with ErrNoAccount error",
					func(firstPage apiclient.AccountPage) {
						accounts := accountClient.List(firstPage)
						moreData := accounts.Next()
						Ω(moreData).Should(BeFalse())
						data, err := accounts.Data()
						Ω(err).Should(MatchError(apiclient.ErrNoAccount))
						Ω(data).Should(BeNil())
						Ω(server.ReceivedRequests()).Should(HaveLen(1))
					},
					Entry("FirstPage", apiclient.FirstPage),
					Entry("default page should behave as first page", apiclient.AccountPage{}),
				)
			})
		})

		Describe("request N-th Page", func() {

			Context("when there is not enught accounts to fill that page", func() {
				var (
					pageNumber int
				)

				BeforeEach(func() {
					responseStatusCode = http.StatusOK // 200
					responseData = &ListAccountResponse{
						Data: nil,
					}
					pageNumber = 10
					rawQuery[0] = "page[number]=10&page[size]=100"
				})

				It("should return nil with ErrNoAccount error", func() {
					accounts := accountClient.List(apiclient.AccountPage{PageNumber: pageNumber})
					moreData := accounts.Next()
					Ω(moreData).Should(BeFalse())
					data, err := accounts.Data()
					Ω(err).Should(MatchError(apiclient.ErrNoAccount))
					Ω(data).Should(BeNil())
					Ω(server.ReceivedRequests()).Should(HaveLen(1))
				})
			})
		})

		Describe("request with filters", func() {
			var (
				listOfAccounts []apiclient.AccountResource
			)

			BeforeEach(func() {
				responseStatusCode = http.StatusOK // 200
				listOfAccounts = libtest.GenerateAccountResources(35)
				responseData = &ListAccountResponse{
					Data: listOfAccounts,
				}
				rawQuery[0] = "page[number]=0&page[size]=100" // default
			})

			DescribeTable("should send request with apropriate query containing filters",
				func(filter apiclient.AccountListFilter, expectedQuery string) {
					if expectedQuery != "" {
						rawQuery[0] = fmt.Sprintf("page[number]=0&page[size]=100&%v", expectedQuery)
					}
					accounts := accountClient.List(apiclient.AccountPage{Filter: filter})
					moreData := accounts.Next()
					Ω(moreData).Should(BeTrue())
					data, err := accounts.Data()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(data).ShouldNot(BeNil())
					Ω(server.ReceivedRequests()).Should(HaveLen(1))
				},
				Entry("empty filter", apiclient.AccountListFilter{}, ""),
				Entry("single AccountNumber", apiclient.AccountListFilter{AccountNumber: []string{"091232190"}}, "filter[account_number]=091232190"),
				Entry("single BankID", apiclient.AccountListFilter{BankID: []string{"34242342"}}, "filter[bank_id]=34242342"),
				Entry("single BankIDCode", apiclient.AccountListFilter{BankIDCode: []string{"23423423"}}, "filter[bank_id_code]=23423423"),
				Entry("single Country", apiclient.AccountListFilter{Country: []string{"GB"}}, "filter[country]=GB"),
				Entry("single CustomerID", apiclient.AccountListFilter{CustomerID: []string{"1322132"}}, "filter[customer_id]=1322132"),
				Entry("single IBAN", apiclient.AccountListFilter{IBAN: []string{"56456464"}}, "filter[iban]=56456464"),
				Entry("multiple AccountNumber", apiclient.AccountListFilter{AccountNumber: []string{"091232190", "9898"}}, "filter[account_number]=091232190,9898"),
				Entry("multiple BankID", apiclient.AccountListFilter{BankID: []string{"34242342", "76766"}}, "filter[bank_id]=34242342,76766"),
				Entry("multiple BankIDCode", apiclient.AccountListFilter{BankIDCode: []string{"23423423", "6645645564"}}, "filter[bank_id_code]=23423423,6645645564"),
				Entry("multiple Country", apiclient.AccountListFilter{Country: []string{"GB", "AU"}}, "filter[country]=GB,AU"),
				Entry("multiple CustomerID", apiclient.AccountListFilter{CustomerID: []string{"1322132", "6546546"}}, "filter[customer_id]=1322132,6546546"),
				Entry("multiple IBAN", apiclient.AccountListFilter{IBAN: []string{"56456464", "123213"}}, "filter[iban]=56456464,123213"),
				Entry("mixture", apiclient.AccountListFilter{IBAN: []string{"56456464", "123213"}, Country: []string{"GB", "AU"}, BankIDCode: []string{"23423423"}}, "filter[country]=GB,AU&filter[iban]=56456464,123213&filter[bank_id_code]=23423423"),
			)
		})
	})

	Describe("List Accounts: request multiple pages", func() {
		var (
			// Pages from 10 to 19 are full, i.e. 37 accounts each
			// Page 20 is not full, i.e. less than 37 accounts
			// Page 21 is empty
			firstPage int = 10
			lastPage  int = 21
			pageSize  int = 37
			accNum    int

			accountsByRequest map[int]([]apiclient.AccountResource)
		)

		BeforeEach(func() {
			accountsByRequest = make(map[int][]apiclient.AccountResource, lastPage-firstPage+1)
			for i := firstPage; i <= lastPage; i += 1 {
				var accountResources []apiclient.AccountResource
				if i < lastPage-1 {
					accountResources = libtest.GenerateAccountResources(pageSize)
				} else if i == lastPage-1 {
					accountResources = libtest.GenerateAccountResources(1 + rand.Intn(pageSize-1))
				} else {
					accountResources = nil
				}
				accountsByRequest[i] = accountResources
				responseData := ListAccountResponse{Data: accountResources}

				rawQuery := fmt.Sprintf("page[number]=%v&page[size]=%v", i, pageSize)
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", listPath, rawQuery),
						ghttp.RespondWithJSONEncoded(200, responseData),
					),
				)
			}
			accNum = (lastPage-firstPage-1)*pageSize + len(accountsByRequest[lastPage-1])
		})

		Context("when using List and then Next only", func() {

			It("should get ten full list, and one not full", func() {
				startPage := apiclient.AccountPage{PageNumber: firstPage, PageSize: pageSize}
				accounts := accountClient.List(startPage)
				pageCount := 0
				accCount := 0
				for accounts.Next() {
					pageCount += 1
					data, err := accounts.Data()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(data).ShouldNot(BeNil())
					if pageCount < lastPage-firstPage {
						Ω(data).Should(HaveLen(pageSize))
					} else {
						Ω(len(data)).Should(BeNumerically("<", pageSize))
					}
					accCount += len(data)
				}

				Ω(pageCount).Should(Equal(lastPage - firstPage))
				Ω(accCount).Should(Equal(accNum))

				Ω(server.ReceivedRequests()).Should(HaveLen(pageCount + 1))
			})
		})

		Context("when using List and then FetchAll only", func() {
			It("should get all elements one by one", func() {
				startPage := apiclient.AccountPage{PageNumber: firstPage, PageSize: pageSize}
				accounts := accountClient.List(startPage)
				accCount := 0
				for range accounts.FetchAll() {
					accCount += 1
				}

				Ω(accCount).Should(Equal(accNum))
			})
		})
	})
})
