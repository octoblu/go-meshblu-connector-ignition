package updateconnector_test

import (
	"github.com/octoblu/go-meshblu-connector-ignition/runner"
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpdateConnector", func() {
	var sut updateconnector.UpdateConnector

	config := &runner.Config{
		ServiceName:   "TestConnector",
		DisplayName:   "TestConnector DisplayName",
		Description:   "TestConnector Description",
		ConnectorName: "test",
		GithubSlug:    "testblu/meshblu-connector-test",
		Tag:           "v1.0.0",
		BinPath:       "path/to/bin",
		Dir:           "path/to/connector",
		Stderr:        "path/to/stderr",
		Stdout:        "path/to/stdout",
	}

	Describe("->New", func() {
		It("should produce an instance", func() {
			Expect(updateconnector.New(config)).NotTo(BeNil())
		})
	})

	Describe("with an instance", func() {
		var err error
		BeforeEach(func() {
			sut, err = updateconnector.New(config)
		})

		It("should not have error", func() {
			Expect(err).To(BeNil())
		})

		Describe("sut.NeedsUpdate", func() {
			Describe("when called", func() {
				var needsUpdate bool
				var err error
				BeforeEach(func() {
					needsUpdate, err = sut.NeedsUpdate()
				})

				It("should not have error", func() {
					Expect(err).To(BeNil())
				})

				It("should need update", func() {
					Expect(needsUpdate).To(BeTrue())
				})
			})
		})
	})
})
