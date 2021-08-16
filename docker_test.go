package main

import (
	"github.com/aserto-dev/go-lib/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("registry-versions", func() {
	Context("image doesn't exist yet", func() {
		It("tags returns an empty array", func() {
			tags, err := imageTags("ghcr.io/aserto-dev/bb9d692b24ad4", "aserto-bot", testutil.VaultValue(ginkgoT, "github-bot.root-pat"))

			Expect(err).ToNot(HaveOccurred())
			Expect(tags).To(HaveLen(0))
		})
	})

	Context("when it's a development release", func() {
		It("only returns the full version tag", func() {
			version := "1.0.0-dev"
			existingTags := []string{}

			tags, err := calculateTagsForVersion(version, existingTags)
			Expect(err).ToNot(HaveOccurred())

			Expect(tags).To(HaveLen(1))
			Expect(tags).To(ContainElement(version))
		})
	})

	Context("when it's not a development release", func() {
		Context("there are no remote tags", func() {
			It("returns the full version, major, minor and latest tags", func() {
				version := "1.0.1"
				existingTags := []string{}

				tags, err := calculateTagsForVersion(version, existingTags)
				Expect(err).ToNot(HaveOccurred())

				Expect(tags).To(HaveLen(4))
				Expect(tags).To(ContainElement("1.0.1"))
				Expect(tags).To(ContainElement("1.0"))
				Expect(tags).To(ContainElement("1"))
				Expect(tags).To(ContainElement("latest"))
			})
		})

		Context("if it's not the latest version", func() {
			It("only returns the full version, and major and minor versions", func() {
				version := "1.3.1"
				existingTags := []string{"0.9.0", "1.2.0", "1.0.0", "2.0.0"}

				tags, err := calculateTagsForVersion(version, existingTags)
				Expect(err).ToNot(HaveOccurred())

				Expect(tags).To(HaveLen(3))
				Expect(tags).To(ContainElement("1.3.1"))
				Expect(tags).To(ContainElement("1.3"))
				Expect(tags).To(ContainElement("1"))
			})

			Context("if it's not the latest in the major series", func() {
				It("only returns the full version, and minor versions", func() {
					version := "1.3.1"
					existingTags := []string{"1.2.0", "1.4.0"}

					tags, err := calculateTagsForVersion(version, existingTags)
					Expect(err).ToNot(HaveOccurred())

					Expect(tags).To(HaveLen(2))
					Expect(tags).To(ContainElement("1.3.1"))
					Expect(tags).To(ContainElement("1.3"))
				})

				Context("but if it's the only one in the major series", func() {
					It("it returns the full version, major and minor versions", func() {
						version := "0.3.1"
						existingTags := []string{"1.2.0", "1.4.0"}

						tags, err := calculateTagsForVersion(version, existingTags)
						Expect(err).ToNot(HaveOccurred())

						Expect(tags).To(HaveLen(3))
						Expect(tags).To(ContainElement("0.3.1"))
						Expect(tags).To(ContainElement("0.3"))
						Expect(tags).To(ContainElement("0"))
					})
				})
			})

			Context("if it's not the latest in the minor series", func() {
				It("only returns the full version", func() {
					version := "1.2.1"
					existingTags := []string{"1.2.2", "1.4.0"}

					tags, err := calculateTagsForVersion(version, existingTags)
					Expect(err).ToNot(HaveOccurred())

					Expect(tags).To(HaveLen(1))
					Expect(tags).To(ContainElement("1.2.1"))
				})
			})
		})

		Context("if it's the latest version", func() {
			It("returns the full version, major, minor and latest tags", func() {
				version := "2.1.1"
				existingTags := []string{"2.1.0", "1.2.0", "2.0.0"}

				tags, err := calculateTagsForVersion(version, existingTags)
				Expect(err).ToNot(HaveOccurred())

				Expect(tags).To(HaveLen(4))
				Expect(tags).To(ContainElement("2.1.1"))
				Expect(tags).To(ContainElement("2.1"))
				Expect(tags).To(ContainElement("2"))
				Expect(tags).To(ContainElement("latest"))
			})
		})
	})
})
