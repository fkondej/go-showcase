package apiclient_test

import (
	"testing"

	"github.com/fkondej/go-showcase/v1/pkg/libtest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestApiclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "apiclient")
}

var _ = BeforeSuite(func() {
	libtest.DBBeforeSuite()
})

var _ = AfterSuite(func() {
	libtest.DBAfterSuite()
})
