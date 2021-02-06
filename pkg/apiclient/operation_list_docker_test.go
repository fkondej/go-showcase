package apiclient_test

import (
	"github.com/fkondej/go-showcase/v1/pkg/apiclient"
	"github.com/fkondej/go-showcase/v1/pkg/libtest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("The AccountClient", func() {
	var (
		accountClient *apiclient.AccountClient
	)

	BeforeEach(func() {
		accountClient = apiclient.NewAccountClient(&DefaultTestConfig)
	})

	Describe("List Account operation", func() {

		Context("when server has multiple accounts", func() {
			var (
				accNum int
			)

			BeforeEach(func() {
				accNum = 330
				libtest.DBDeleteAllAccounts()
				libtest.DBCreateAccounts(accNum)
			})

			It("should return non empty first page", func() {
				accounts := accountClient.List(apiclient.FirstPage)

				moreResults := accounts.Next()
				Ω(moreResults).Should(BeTrue())

				data, err := accounts.Data()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(data).ShouldNot(BeNil())
			})

			It("should return an empty page at some point", func() {
				accounts := accountClient.List(apiclient.FirstPage)

				var pageCount = 0
				accCount := 0
				for accounts.Next() {
					pageCount += 1
					data, err := accounts.Data()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(data).ShouldNot(BeNil())

					accCount += len(data)

					if pageCount > 5 {
						Fail("Cannot reach the last/empty page")
					}
				}

				Ω(pageCount).Should(Equal(4))
				Ω(accCount).Should(Equal(accNum))
			})

			Context("when using List and then FetchAll only", func() {
				It("should get all elements one by one", func() {
					startPage := apiclient.AccountPage{PageNumber: 0, PageSize: 13}
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
})
