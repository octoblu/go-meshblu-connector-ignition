package updateconnector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUpdateConnector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UpdateConnector Suite")
}
