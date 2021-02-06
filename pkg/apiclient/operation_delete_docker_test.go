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

	Describe("Delete Account operation", func() {

		Context("when the Account exists", func() {
			var (
				accountID string
			)

			BeforeEach(func() {
				dbAccount := libtest.DBCreateAccounts(1)[0]
				accountID = dbAccount.ID.String()
			})

			It("should send request to delete account and inform about successful deletion", func() {
				deleted, err := accountClient.Delete(accountID, 0)
				Ω(deleted).Should(BeTrue())
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
