package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("Config", func() {
	It("returns correct deploymentFile when deployment is set", func() {
		c := Config{Deployment: "/fake-path/manifest.yml"}
		Expect(c.DeploymentFile()).To(Equal("/fake-path/deployment.json"))
	})

	Describe("Paths", func() {
		var c Config
		BeforeEach(func() {
			c = Config{
				ContainingDir:  "/home/fake",
				DeploymentUUID: "madcow",
			}
		})

		It("returns the blobstore path", func() {
			Expect(c.BlobstorePath()).To(Equal("/home/fake/.bosh_micro/madcow/blobs"))
		})

		It("returns the index path", func() {
			Expect(c.IndexPath()).To(Equal("/home/fake/.bosh_micro/madcow/index.json"))
		})

		It("returns the packages path", func() {
			Expect(c.PackagesPath()).To(Equal("/home/fake/.bosh_micro/madcow/packages"))
		})
	})
})