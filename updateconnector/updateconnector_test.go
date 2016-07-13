package updateconnector_test

import (
	"github.com/octoblu/go-meshblu-connector-ignition/runner"
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
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
		fs := afero.NewMemMapFs()
		It("should produce an instance", func() {
			Expect(updateconnector.New(config, fs)).NotTo(BeNil())
		})
	})

	Describe("with an existing update config", func() {
		var err error
		fs := afero.NewMemMapFs()
		updateConfig, updateConfigErr := updateconnector.NewUpdateConfig(fs)

		It("should not have update config err", func() {
			Expect(updateConfigErr).ToNot(HaveOccurred())
		})

		BeforeEach(func() {
			err = updateConfig.Write("v1.0.0")
			if err != nil {
				return
			}
			sut, err = updateconnector.New(config, fs)
		})

		It("should not have error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("sut.NeedsUpdate", func() {
			Describe("when called", func() {
				var needsUpdate bool
				var err error
				BeforeEach(func() {
					needsUpdate, err = sut.NeedsUpdate("v1.0.0")
				})

				It("should not have error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("should not require an update", func() {
					Expect(needsUpdate).To(BeFalse())
				})
			})
		})
	})

	Describe("with an new instance and no update config", func() {
		var err error
		fs := afero.NewMemMapFs()

		BeforeEach(func() {
			sut, err = updateconnector.New(config, fs)
		})

		It("should not have error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("sut.NeedsUpdate", func() {
			Describe("when called", func() {
				var needsUpdate bool
				var err error

				BeforeEach(func() {
					needsUpdate, err = sut.NeedsUpdate("v1.0.0")
				})

				It("should not have error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("should need update", func() {
					Expect(needsUpdate).To(BeTrue())
				})
			})
		})
	})
})
