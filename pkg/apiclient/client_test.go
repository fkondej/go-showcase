package apiclient_test

import (
	"os"

	"github.com/fkondej/go-showcase/v1/pkg/apiclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	DefaultTestConfig = apiclient.AccountClientConfig{
		URL:      getTestServerURL(),
		ProxyURL: nil,
	}
)

func getTestServerURL() string {
	serverURL := "http://localhost:8080/v1/account/"
	if val, ok := os.LookupEnv("SERVERAPI_URL"); ok {
		serverURL = val
	}
	return serverURL
}

var _ = Describe("Exported functionality", func() {
	Context("NewAccountClient", func() {
		It("should not be nil", func() {
			Ω(apiclient.NewAccountClient(&DefaultTestConfig)).ShouldNot(BeNil())
		})
	})
})
