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

	Describe("Fetch Account Information operation", func() {

		Context("when Account exists", func() {
			var (
				dbAccount *libtest.DBAccount
				accountID string
			)

			BeforeEach(func() {
				dbAccount = libtest.DBCreateAccounts(1)[0]
				accountID = dbAccount.ID.String()
			})

			It("should return Account information without error", func() {

				accountInfo, err := accountClient.Fetch(accountID)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(accountInfo).ShouldNot(BeNil())

				Ω(accountInfo.Type).Should(Equal("account"))
				Ω(accountInfo.ID).Should(Equal(accountID))
				Ω(accountInfo.OrganisationID).Should(Equal(dbAccount.OrganisationID.String()))

				Ω(accountInfo.Attributes.Country).Should(Equal(dbAccount.Record.Country))
				Ω(accountInfo.Attributes.BaseCurrency).Should(Equal(dbAccount.Record.BaseCurrency))
				Ω(accountInfo.Attributes.BankID).Should(Equal(dbAccount.Record.BankID))
				Ω(accountInfo.Attributes.BankIDCode).Should(Equal(dbAccount.Record.BankIDCode))
				Ω(accountInfo.Attributes.BIC).Should(Equal(dbAccount.Record.BIC))
				// TODO: there is a problem with accountapi: multiple fileds are not set or retrived, e.g. name, status
			})
		})

		Context("when Account does not exist", func() {
			var (
				accountID string
			)

			BeforeEach(func() {
				accountID = libtest.GenerateID()
				libtest.DBDeleteAccount(accountID)
			})

			It("should return empty result with ErrNoAccount error", func() {
				accountInfo, err := accountClient.Fetch(accountID)
				Ω(err).Should(MatchError(apiclient.ErrNoAccount))
				Ω(accountInfo).Should(BeNil())
			})
		})
	})
})
