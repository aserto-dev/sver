package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("registry-versions", func() {

	Context("when it's a development release", func() {
		It("only returns the full version tag", func() {
			version := "1.0.0-dev"
			existringTags := []string{}

			tags, err := calculateTagsForVersion(version, existringTags)
			Expect(err).ToNot(HaveOccurred())

			Expect(tags).To(HaveLen(1))
			Expect(tags).To(ContainElement(version))
		})
	})

	Context("when it's not a development release", func() {
		Context("there are no remote tags", func() {
			It("returns the full version, major, minor and latest tags", func() {
				version := "1.0.1"
				existringTags := []string{}

				tags, err := calculateTagsForVersion(version, existringTags)
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
				version := "1.0.1"
				existringTags := []string{"0.9.0", "1.2.0", "1.0.0"}

				tags, err := calculateTagsForVersion(version, existringTags)
				Expect(err).ToNot(HaveOccurred())

				Expect(tags).To(HaveLen(3))
				Expect(tags).To(ContainElement("1.0.1"))
				Expect(tags).To(ContainElement("1.0"))
				Expect(tags).To(ContainElement("1"))
			})
		})

		Context("if it's the latest version", func() {
			It("returns the full version, major, minor and latest tags", func() {
				version := "2.1.1"
				existringTags := []string{"2.1.0", "1.2.0", "2.0.0"}

				tags, err := calculateTagsForVersion(version, existringTags)
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
