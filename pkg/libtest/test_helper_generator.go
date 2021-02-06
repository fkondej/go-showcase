package libtest

import (
	"math/rand"
	"time"

	"github.com/fkondej/go-showcase/v1/pkg/apiclient"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func GenerateAccountAttributes() *apiclient.AccountAttributes {
	countryCode := RandomCountryCode()
	currency := RandomCurrency()
	bic := GenerateBIC(countryCode)
	status := RandomStatus()
	result := apiclient.AccountAttributes{
		Country:      countryCode,
		BaseCurrency: &currency,
		BIC:          &bic,
		Name:         GenerateName(),
		Status:       &status,
	}

	bankID := GenerateBankID(countryCode)
	// Not every country supports this, e.g. Netherlands
	if bankID != "" {
		result.BankID = &bankID
	}
	bankIDCode := countryCodeToBankIDCode[countryCode]
	// Not every country supports this, e.g. Netherlands
	if bankIDCode != "" {
		result.BankIDCode = &bankIDCode
	}

	return &result
}

func GenerateAccountResources(num int) []apiclient.AccountResource {
	result := make([]apiclient.AccountResource, num)

	for i := 0; i < num; i += 1 {
		result[0] = apiclient.AccountResource{
			Type:           "accounts",
			ID:             GenerateID(),
			OrganisationID: GenerateOrganisationID(),
			Version:        rand.Intn(999),
			Attributes:     GenerateAccountAttributes(),
		}
	}

	return result
}

var seededRand = func() *rand.Rand {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)

	seededRand := rand.New(rand.NewSource(seed))
	uuid.SetRand(seededRand)
	return seededRand
}()

func GenerateID() string {
	newUUID, err := uuid.NewRandom()
	Ω(err).ShouldNot(HaveOccurred())

	return newUUID.String()
}

func GenerateOrganisationID() string {
	newUUID, err := uuid.NewRandom()
	Ω(err).ShouldNot(HaveOccurred())

	return newUUID.String()
}

var countryCodeList = []string{
	"GB", // United Kingdom
	"AU", // Astralia
	"BE", // Belgium
	"CA", // Canada
	"FR", // France
	"DE", // Germany
	"GR", // Greece
	"HK", // Hong Kong
	"IT", // Italy
	"LU", // Luxembourg
	"NL", // Netherlands
	"PL", // Poland
	"PT", // Portugal
	"ES", // Spain
	"CH", // Switzerland
	"US", // United States
}

func RandomCountryCode() string {
	return countryCodeList[rand.Intn(len(countryCodeList))]
}

var currencyList = []string{
	"GBP", // Pound sterling
	"EUR", // Euro
	"AUD", // Australian dollar
	"CAD", // Canadian dollar
	"HKD", // Hong Kong dollar
	"PLN", // Polish złoty
	"CHF", // Swiss franc
	"USD", // United States dollar
}

func RandomCurrency() string {
	return currencyList[rand.Intn(len(currencyList))]
}

func RandomStatus() string {
	statusList := []string{"confirmed", "pending"}
	return statusList[rand.Intn(len(statusList))]
}

const asciiLowercase = "abcdefghijklmnopqrstuvwxyz"
const asciiUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const digits = "0123456789"

func GenerateName() [4]string {
	var randNameWord = func() string {
		size := 3 + rand.Intn(25)
		buf := make([]byte, size)
		buf[0] = asciiUppercase[rand.Intn(len(asciiUppercase))]
		for i := 1; i < size; i += 1 {
			buf[i] = asciiLowercase[rand.Intn(len(asciiLowercase))]
		}
		return string(buf)
	}
	result := [4]string{"", "", "", ""}
	result[0] = randNameWord() + " " + randNameWord() + " " + randNameWord()
	result[1] = randNameWord()
	return result
}

var countryCodeToBankIDCode = map[string]string{
	"GB": "GBDSC", // United Kingdom
	"AU": "AUBSB", // Astralia
	"BE": "BE",    // Belgium
	"CA": "CACPA", // Canada
	"FR": "FR",    // France
	"DE": "DEBLZ", // Germany
	"GR": "GRBIC", // Greece
	"HK": "HKNCC", // Hong Kong
	"IT": "ITNCC", // Italy
	"LU": "LULUX", // Luxembourg
	"NL": "",      // Netherlands
	"PL": "PLKNR", // Poland
	"PT": "PTNCC", // Portugal
	"ES": "ESNCC", // Spain
	"CH": "CHBCC", // Switzerland
	"US": "USABA", // United States
}

func GenerateBIC(countryCode string) string {
	// SWIFT BIC in either 8 or 11 character format e.g. 'NWBKGB22'
	var size int = 8
	if rand.Intn(2) == 0 {
		size = 11
	}
	buf := make([]byte, size)

	for i := 0; i < 4; i += 1 {
		buf[i] = asciiUppercase[rand.Intn(len(asciiUppercase))]
	}
	buf[4] = countryCode[0]
	buf[5] = countryCode[1]

	for i := 6; i < size; i += 1 {
		buf[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return string(buf)
}

func GenerateBankID(countryCode string) string {
	ccToSize := map[string]int{
		"GB": 6,  // United Kingdom
		"AU": 6,  // Astralia
		"BE": 3,  // Belgium
		"CA": 9,  // Canada
		"FR": 10, // France
		"DE": 8,  // Germany
		"GR": 7,  // Greece
		"HK": 3,  // Hong Kong
		"IT": 11, // Italy
		"LU": 3,  // Luxembourg
		"NL": 0,  // Netherlands
		"PL": 8,  // Poland
		"PT": 8,  // Portugal
		"ES": 8,  // Spain
		"CH": 5,  // Switzerland
		"US": 9,  // United States
	}
	size := ccToSize[countryCode]
	buf := make([]byte, size)
	for i := 0; i < size; i += 1 {
		buf[i] = digits[rand.Intn(len(digits))]

	}
	return string(buf)
}

var _ = Describe("Generators", func() {

	Context("when GenerateID is called", func() {

		It("should generate new ID of the same length", func() {
			for i := 1; i <= 10; i += 1 {
				accountID := GenerateID()
				Ω(accountID).ShouldNot(BeNil())
				Ω(accountID).Should(HaveLen(36))
			}
		})

		It("should generate different values", func() {
			for i := 1; i <= 3; i += 1 {
				if GenerateID() != GenerateID() {
					return
				}
			}
			Fail("Generated the same value")
		})
	})

	Context("when GenerateOrganisationID is called", func() {

		It("should generate new OrganisationID of the same length", func() {
			for i := 1; i <= 10; i += 1 {
				organisationID := GenerateOrganisationID()
				Ω(organisationID).ShouldNot(BeNil())
				Ω(organisationID).Should(HaveLen(36))
			}
		})

		It("should generate different values", func() {
			for i := 1; i <= 3; i += 1 {
				if GenerateOrganisationID() != GenerateOrganisationID() {
					return
				}
			}
			Fail("Generated the same value")
		})
	})

	Context("when RandomCountryCode is called", func() {

		It("should randomly select a country code", func() {
			for i := 1; i <= 10; i += 1 {
				countryCode := RandomCountryCode()
				Ω(countryCode).ShouldNot(BeNil())
				Ω(countryCode).Should(HaveLen(2))
			}
		})

		It("should random different values", func() {
			for i := 1; i <= 3; i += 1 {
				if RandomCountryCode() != RandomCountryCode() {
					return
				}
			}
			Fail("Generated the same value")
		})
	})

	Context("when RandomCurrency is called", func() {

		It("should randomly select a currency", func() {
			for i := 1; i <= 10; i += 1 {
				currency := RandomCurrency()
				Ω(currency).ShouldNot(BeNil())
				Ω(currency).Should(HaveLen(3))
			}
		})

		It("should random different values", func() {
			for i := 1; i <= 10; i += 1 {
				if RandomCurrency() != RandomCurrency() {
					return
				}
			}
			Fail("Generated the same value")
		})
	})

	Context("when RandomStatus is called", func() {

		It("should randomly select an account status", func() {
			for i := 1; i <= 10; i += 1 {
				status := RandomStatus()
				Ω(status).ShouldNot(BeNil())
				Ω(status).Should(BeElementOf("confirmed", "pending"))
			}
		})

		It("should random different values", func() {
			for i := 1; i <= 10; i += 1 {
				if RandomStatus() != RandomStatus() {
					return
				}
			}
			Fail("Generated the same value")
		})
	})

	Context("when GenerateName is called", func() {

		It("should generate user name", func() {
			for i := 1; i <= 10; i += 1 {
				name := GenerateName()
				Ω(name).ShouldNot(BeNil())
				Ω(name).Should(HaveLen(4))
				Ω(len(name[0])).Should(BeNumerically(">", 10))
			}
		})

		It("should random different values", func() {
			for i := 1; i <= 3; i += 1 {
				if GenerateName()[0] != GenerateName()[0] {
					return
				}
			}
			Fail("Generated the same value")
		})
	})

	Context("when GenerateBIC is called", func() {

		It("should generate bank BIC number", func() {
			for _, countryCode := range countryCodeList {
				bic := GenerateBIC(countryCode)
				Ω(bic).ShouldNot(BeNil())
				Ω(len(bic)).Should(BeElementOf(8, 11))
			}
		})

		It("should random different values", func() {
			for _, countryCode := range countryCodeList {
				if GenerateBIC(countryCode) != GenerateBIC(countryCode) {
					return
				}
			}
			Fail("Generated the same value")
		})
	})

	Context("when GenerateBankID is called", func() {

		It("should generate bank id", func() {
			for _, countryCode := range countryCodeList {
				bankID := GenerateBankID(countryCode)
				Ω(bankID).ShouldNot(BeNil())
			}
		})

		It("should random different values", func() {
			for _, countryCode := range countryCodeList {
				if GenerateBankID(countryCode) != GenerateBankID(countryCode) {
					return
				}
			}
			Fail("Generated the same value")
		})
	})

	Context("when GenerateAccountAttributes is called", func() {

		It("should generate Account Attributes", func() {
			for i := 1; i <= 10; i += 1 {
				accountAttributes := GenerateAccountAttributes()
				Ω(accountAttributes).ShouldNot(BeNil())
				Ω(accountAttributes.Country).ShouldNot(BeNil())
				Ω(accountAttributes.BaseCurrency).ShouldNot(BeNil())
				Ω(accountAttributes.BIC).ShouldNot(BeNil())
				Ω(accountAttributes.Name).ShouldNot(BeNil())
				Ω(accountAttributes.Status).ShouldNot(BeNil())
				Ω(accountAttributes.BankID).ShouldNot(Equal(""))
				Ω(accountAttributes.BankIDCode).ShouldNot(Equal(""))
			}
		})

		It("should random different values", func() {
			for i := 1; i < 3; i += 1 {
				if GenerateAccountAttributes() != GenerateAccountAttributes() {
					return
				}
			}
			Fail("Generated the same value")
		})
	})

})
