package updateconnector_test

import (
	"encoding/json"
	"path/filepath"

	"github.com/octoblu/go-meshblu-connector-ignition/logger"
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"

	"github.com/kardianos/osext"
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
			sut, err = updateconnector.New("testblu/test", "test", "path/to/dir", fs, logger.NewFakeMainLogger())
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
		BeforeEach(func() {
			fullPath, _ := osext.Executable()
			dir, _ := filepath.Split(fullPath)
			filePath := filepath.Join(dir, "package.json")
			type packageJSON struct {
				Version string `json:"version"`
			}
			packageConfig := &packageJSON{
				Version: "1.0.0",
			}
			jsonBytes, _ := json.Marshal(packageConfig)
			afero.WriteFile(fs, filePath, jsonBytes, 0644)
		})

		BeforeEach(func() {
			sut, err = updateconnector.New("testblu/test", "test", "path/to/dir", fs, logger.NewFakeMainLogger())
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

		BeforeEach(func() {
			fullPath, _ := osext.Executable()
			dir, _ := filepath.Split(fullPath)
			filePath := filepath.Join(dir, "package.json")
			type packageJSON struct {
				Version string `json:"version"`
			}
			packageConfig := &packageJSON{
				Version: "1.10.0",
			}
			jsonBytes, _ := json.Marshal(packageConfig)
			afero.WriteFile(fs, filePath, jsonBytes, 0644)
		})

		BeforeEach(func() {
			sut, err = updateconnector.New("testblu/test", "test", "path/to/dir", fs, logger.NewFakeMainLogger())
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
			sut, err = updateconnector.New("testblu/test", "test", "path/to/dir", fs, logger.NewFakeMainLogger())
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

				It("should require an update", func() {
					Expect(needsUpdate).To(BeTrue())
				})
			})
		})
	})
})
