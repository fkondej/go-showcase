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

	Describe("Create Account operation", func() {
		var (
			accountID         string
			organisationID    string
			accountAttributes *apiclient.AccountAttributes
		)

		BeforeEach(func() {
			accountID = libtest.GenerateID()
			organisationID = libtest.GenerateOrganisationID()
			accountAttributes = libtest.GenerateAccountAttributes()
		})

		Context("when Account does not exist", func() {

			BeforeEach(func() {
				// make sure account with given ID does not exists
				libtest.DBDeleteAccount(accountID)
			})

			It("should successfully create an account", func() {
				createdAccount, err := accountClient.Create(accountID, organisationID, accountAttributes)

				Ω(err).ShouldNot(HaveOccurred())

				Ω(createdAccount).ShouldNot(BeNil())
			})
		})
	})
})
