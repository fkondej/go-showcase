package apiclient

// Only example unit test
// it is within 'apiclient' package in order to access non public functions to test them closely

import (
	"fmt"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("operation List functionality", func() {

	Describe("convertToQuery", func() {

		Context("without filters", func() {
			It("should not add filters to query", func() {
				Ω(
					convertToQuery(0, 100, AccountListFilter{}),
				).Should(Equal(
					url.Values{
						"page[number]": []string{"0"},
						"page[size]":   []string{"100"},
					},
				))
			})
		})

		Context("with filters", func() {
			It("should add filters to query", func() {
				Ω(
					convertToQuery(0, 100, AccountListFilter{
						AccountNumber: []string{"897897", "342434"},
						BankID:        []string{"8768768", "13123213"},
						BankIDCode:    []string{"453543", "787686"},
						Country:       []string{"UK", "US", "AU"},
						CustomerID:    []string{"123", "56646"},
						IBAN:          []string{"7876868", "432432"},
					}),
				).Should(Equal(
					url.Values{
						"page[number]":           []string{"0"},
						"page[size]":             []string{"100"},
						"filter[account_number]": []string{"897897,342434"},
						"filter[bank_id]":        []string{"8768768,13123213"},
						"filter[bank_id_code]":   []string{"453543,787686"},
						"filter[country]":        []string{"UK,US,AU"},
						"filter[customer_id]":    []string{"123,56646"},
						"filter[iban]":           []string{"7876868,432432"},
					},
				))
			})
		})
	})

	Describe("AccountPageResult", func() {

		Context("when called Data before Next", func() {

			It("should return error", func() {
				accounts := AccountPageResult{}

				data, err := accounts.Data()
				Ω(err).Should(HaveOccurred())
				Ω(data).Should(BeNil())
			})
		})

		Context("when called Next after it encountered error", func() {

			It("should return false", func() {
				accounts := AccountPageResult{
					lastError: fmt.Errorf("Error"),
				}
				Ω(accounts.Next()).Should(BeFalse())
			})
		})
	})
})
