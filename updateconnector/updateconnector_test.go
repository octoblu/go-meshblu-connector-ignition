package updateconnector_test

import (
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("UpdateConnector", func() {
	var sut updateconnector.UpdateConnector

	Describe("->New", func() {
		var err error
		fs := afero.NewMemMapFs()
		BeforeEach(func() {
			sut, err = updateconnector.New("testblu/test", "test", "path/to/dir", fs)
		})

		It("should not return a error", func() {
			Expect(err).To(BeNil())
		})

		It("should produce an instance", func() {
			Expect(sut).NotTo(BeNil())
		})
	})

	Describe("with an existing config of the current version", func() {
		var err error
		fs := afero.NewMemMapFs()
		updateConfig, updateConfigErr := updateconnector.NewUpdateConfig(fs)

		It("should not have update config err", func() {
			Expect(updateConfigErr).To(BeNil())
		})

		BeforeEach(func() {
			err = updateConfig.Write("v1.0.0", 0)
			if err != nil {
				return
			}
			sut, err = updateconnector.New("testblu/test", "test", "path/to/dir", fs)
		})

		It("should not have error", func() {
			Expect(err).To(BeNil())
		})

		Describe("sut.NeedsUpdate", func() {
			Describe("when called", func() {
				var needsUpdate bool
				var err error
				BeforeEach(func() {
					needsUpdate, err = sut.NeedsUpdate("v1.0.0")
				})

				It("should not have error", func() {
					Expect(err).To(BeNil())
				})

				It("should not require an update", func() {
					Expect(needsUpdate).To(BeFalse())
				})
			})
		})
	})

	Describe("with an existing config of the wrong version", func() {
		var err error
		fs := afero.NewMemMapFs()
		updateConfig, updateConfigErr := updateconnector.NewUpdateConfig(fs)

		It("should not have update config err", func() {
			Expect(updateConfigErr).To(BeNil())
		})

		BeforeEach(func() {
			err = updateConfig.Write("v1.5.0", 0)
			if err != nil {
				return
			}
			sut, err = updateconnector.New("testblu/test", "test", "path/to/dir", fs)
		})

		It("should not have error", func() {
			Expect(err).To(BeNil())
		})

		Describe("sut.NeedsUpdate", func() {
			Describe("when called", func() {
				var needsUpdate bool
				var err error
				BeforeEach(func() {
					needsUpdate, err = sut.NeedsUpdate("v1.3.0")
				})

				It("should not have error", func() {
					Expect(err).To(BeNil())
				})

				It("should require an update", func() {
					Expect(needsUpdate).To(BeTrue())
				})
			})
		})
	})

	Describe("with an new instance and no update config", func() {
		var err error
		fs := afero.NewMemMapFs()

		BeforeEach(func() {
			sut, err = updateconnector.New("testblu/test", "test", "path/to/dir", fs)
		})

		It("should not have error", func() {
			Expect(err).To(BeNil())
		})

		Describe("sut.NeedsUpdate", func() {
			Describe("when called", func() {
				var needsUpdate bool
				var err error

				BeforeEach(func() {
					needsUpdate, err = sut.NeedsUpdate("v1.0.0")
				})

				It("should not have error", func() {
					Expect(err).To(BeNil())
				})

				It("should requrie an update", func() {
					Expect(needsUpdate).To(BeTrue())
				})
			})
		})
	})
})
