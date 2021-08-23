package sver

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("sver", func() {

	var (
		dir string
		err error
	)

	BeforeEach(func() {
		dir, err = ioutil.TempDir("", "sver")
		Expect(err).ToNot(HaveOccurred())
		err = os.Chdir(dir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(dir)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe(".current_version", func() {
		Context("when git does not exist", func() {

			It("raises an error", func() {
				oldBinary := gitBinary
				gitBinary = "thisbinarydoesnotexist"
				defer func() { gitBinary = oldBinary }()
				err := verifyGit()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when current dir is not a git work tree", func() {
			It("raises an error", func() {
				_, err := CurrentVersion(false, false)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when current dir is a git work tree", func() {
			It("does not raise", func() {
				createGitDirWithTag("v0.0.1")
				_, err := CurrentVersion(false, false)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the current tag is a semver version", func() {
			It("does not raise an error", func() {
				createGitDirWithTag("v0.0.1")
				_, err := CurrentVersion(false, false)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the current tag is a not a semver version", func() {
			It("raises an error", func() {
				createGitDirWithTag("some_tag")
				_, err := CurrentVersion(false, false)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when no git tag exists", func() {
			It("returns an initial version", func() {
				_, err := git("init")
				createCommit("foo")

				Expect(err).ToNot(HaveOccurred())

				version, err := CurrentVersion(false, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(version).To(MatchRegexp(`^0\.0\.0-[0-9]{14}\.0\.g[0-9a-fA-F]{8}$`))
			})
		})

		Context("when the current tag is a semver tag with a `+` element", func() {
			It("raise with an error that this not supported", func() {
				createGitDirWithTag("1.0.2+gold")
				_, err := CurrentVersion(false, false)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the current tag is a semver tag without a `v` in front", func() {
			It("does not raise and returns the correct semver version", func() {
				createGitDirWithTag("1.0.2")
				version, err := CurrentVersion(false, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(version).To(MatchRegexp(`^1\.0\.2$`))
			})
		})

		Context("with newer commits since the current semver tag", func() {
			Context("and a release version", func() {
				BeforeEach(func() {
					createGitDirWithTag("v1.0.2")
					createCommit("test")
				})

				Context("when there are no uncommitted changes", func() {
					It("returns a pre-release version without a dirty tag", func() {
						version, err := CurrentVersion(false, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(version).To(MatchRegexp(`^1\.0\.2-[0-9]{14}\.1\.g[0-9a-fA-F]{8}$`))
					})

					Context("when release flag is present", func() {
						It("returns an error", func() {
							_, err := CurrentVersion(true, false)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("not on a tag, this is a pre release version"))
						})
					})
				})

				Context("when there are uncommitted changes", func() {
					Context("in files tracked by git", func() {
						BeforeEach(func() {
							createUncomittedChanges()
						})

						It("returns a pre-release version with a dirty tag", func() {
							version, err := CurrentVersion(false, false)
							Expect(err).ToNot(HaveOccurred())
							Expect(version).To(MatchRegexp(`^1\.0\.2-[0-9]{14}\.1\.g[0-9a-fA-F]{8}-dirty$`))
						})

						Context("when release flag is present", func() {
							It("returns an error", func() {
								_, err := CurrentVersion(true, false)
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("not on a tag, this is a pre release version"))
							})
						})

						Context("when force flag is present", func() {
							It("returns a version anyway without the dirty flag", func() {
								version, err := CurrentVersion(false, true)
								Expect(err).ToNot(HaveOccurred())
								Expect(version).To(MatchRegexp(`^1\.0\.2-[0-9]{14}\.1\.g[0-9a-fA-F]{8}$`))
							})
						})
					})

					Context("in files not tracked by git", func() {
						It("returns a pre-release version with a dirty tag", func() {
							err := ioutil.WriteFile("some_untracked_file", []byte("Dummy content"), 0600)
							Expect(err).ToNot(HaveOccurred())

							version, err := CurrentVersion(false, false)
							Expect(err).ToNot(HaveOccurred())
							Expect(version).To(MatchRegexp(`^1\.0\.2-[0-9]{14}\.1\.g[0-9a-fA-F]{8}-dirty$`))
						})
					})
				})
			})

			Context("and an alpha version", func() {
				BeforeEach(func() {
					createGitDirWithTag("v2.4.0-alpha.foo")
					createCommit("test")
				})

				Context("when there are no uncommitted changes", func() {
					It("returns a pre-release version without a dirty tag", func() {
						version, err := CurrentVersion(false, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(version).To(MatchRegexp(`^2\.4\.0-alpha\.foo-[0-9]{14}\.1\.g[0-9a-fA-F]{8}$`))
					})
				})

				Context("when there are uncommitted changes", func() {
					Context("in files tracked by git", func() {
						It("returns a pre-release version with a dirty tag", func() {
							createUncomittedChanges()

							version, err := CurrentVersion(false, false)
							Expect(err).ToNot(HaveOccurred())
							Expect(version).To(MatchRegexp(`^2\.4\.0-alpha\.foo-[0-9]{14}\.1\.g[0-9a-fA-F]{8}-dirty$`))
						})
					})

					Context("in files not tracked by git", func() {
						It("returns a pre-release version with a dirty tag", func() {
							err := ioutil.WriteFile("some_untracked_file", []byte("Dummy content"), 0600)
							Expect(err).ToNot(HaveOccurred())

							version, err := CurrentVersion(false, false)
							Expect(err).ToNot(HaveOccurred())
							Expect(version).To(MatchRegexp(`^2\.4\.0-alpha\.foo-[0-9]{14}\.1\.g[0-9a-fA-F]{8}-dirty$`))
						})

						Context("when force flag is present", func() {
							It("returns a version anyway without the dirty flag", func() {
								version, err := CurrentVersion(false, true)
								Expect(err).ToNot(HaveOccurred())
								Expect(version).To(MatchRegexp(`^2\.4\.0-alpha\.foo-[0-9]{14}\.1\.g[0-9a-fA-F]{8}$`))
							})
						})
					})
				})
			})
		})

		Context("with no new commits since the current semver tag", func() {
			Context("and a release version", func() {
				BeforeEach(func() {
					createGitDirWithTag("v1.0.2")
				})

				Context("when there are no uncommitted changes", func() {
					It("returns just the release version", func() {
						version, err := CurrentVersion(false, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(version).To(MatchRegexp(`^1\.0\.2$`))
					})
				})

				Context("when there are uncommitted changes", func() {
					BeforeEach(func() {
						createUncomittedChanges()
					})

					It("returns the release version with a dirty tag", func() {
						version, err := CurrentVersion(false, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(version).To(MatchRegexp(`^1\.0\.2-dirty$`))
					})

					Context("when release flag is present", func() {
						It("returns an error", func() {
							_, err := CurrentVersion(true, false)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("version is dirty"))
						})
					})

					Context("when release and dirty flags are present", func() {
						It("returns a version", func() {
							version, err := CurrentVersion(true, true)
							Expect(err).ToNot(HaveOccurred())
							Expect(version).To(MatchRegexp(`^1\.0\.2$`))
						})
					})
				})
			})

			Context("and an alpha version", func() {
				BeforeEach(func() {
					createGitDirWithTag("v2.4.0-alpha.foo")
				})

				Context("when there are no uncommitted changes", func() {
					It("returns just the alpha version", func() {
						version, err := CurrentVersion(false, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(version).To(MatchRegexp(`^2\.4\.0-alpha\.foo$`))
					})
				})

				Context("when there are uncommitted changes", func() {
					It("returns the alpha version with a dirty tag", func() {
						createUncomittedChanges()
						version, err := CurrentVersion(false, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(version).To(MatchRegexp(`^2\.4\.0-alpha\.foo-dirty$`))
					})
				})
			})
		})
	})

	Describe("pre-release", func() {
		BeforeEach(func() {
			createGitDirWithTag("v10.200.5")
		})

		Context("when pre-release is used", func() {
			It("calculates the next patch version", func() {
				version, err := CurrentVersion(false, false)
				Expect(err).ToNot(HaveOccurred())
				version = PreRelease(version, "nightly")

				Expect(version).To(Equal(`10.200.5-nightly`))
			})
		})

		Context("when pre-release is used and worktree is dirty", func() {
			BeforeEach(func() {
				createUncomittedChanges()
			})

			It("calculates the next patch version", func() {
				version, err := CurrentVersion(false, false)
				Expect(err).ToNot(HaveOccurred())
				version = PreRelease(version, "nightly")

				Expect(version).To(Equal(`10.200.5-dirty-nightly`))
			})
		})
	})

	Describe("next", func() {
		BeforeEach(func() {
			createGitDirWithTag("v10.200.5")
		})

		Context("when patch is used", func() {
			It("calculates the next patch version", func() {
				version, err := CurrentVersion(false, false)
				Expect(err).ToNot(HaveOccurred())
				version, err = Next(version, "patch")
				Expect(err).ToNot(HaveOccurred())

				Expect(version).To(Equal(`10.200.6`))
			})
		})

		Context("when minor is used", func() {
			It("calculates the next minor version", func() {
				version, err := CurrentVersion(false, false)
				Expect(err).ToNot(HaveOccurred())
				version, err = Next(version, "minor")
				Expect(err).ToNot(HaveOccurred())

				Expect(version).To(Equal(`10.201.0`))
			})
		})

		Context("when major is used", func() {
			It("calculates the next major version", func() {
				version, err := CurrentVersion(false, false)
				Expect(err).ToNot(HaveOccurred())
				version, err = Next(version, "major")
				Expect(err).ToNot(HaveOccurred())

				Expect(version).To(Equal(`11.0.0`))
			})
		})

		Context("when the tree is dirty", func() {
			BeforeEach(func() {
				createUncomittedChanges()
			})

			Context("when patch is used", func() {
				It("calculates the next patch version", func() {
					version, err := CurrentVersion(false, false)
					Expect(err).ToNot(HaveOccurred())
					version, err = Next(version, "patch")
					Expect(err).ToNot(HaveOccurred())

					Expect(version).To(Equal(`10.200.6-dirty`))
				})
			})

			Context("when minor is used", func() {
				It("calculates the next minor version", func() {
					version, err := CurrentVersion(false, false)
					Expect(err).ToNot(HaveOccurred())
					version, err = Next(version, "minor")
					Expect(err).ToNot(HaveOccurred())

					Expect(version).To(Equal(`10.201.0-dirty`))
				})
			})

			Context("when major is used", func() {
				It("calculates the next major version", func() {
					version, err := CurrentVersion(false, false)
					Expect(err).ToNot(HaveOccurred())
					version, err = Next(version, "major")
					Expect(err).ToNot(HaveOccurred())

					Expect(version).To(Equal(`11.0.0-dirty`))
				})
			})
		})
	})
})

func createCommit(fileName string) {
	err := ioutil.WriteFile(fileName, []byte("Dummy content"), 0600)
	Expect(err).ToNot(HaveOccurred())

	_, err = git("add", fileName)
	Expect(err).ToNot(HaveOccurred())

	_, err = git("commit", "--no-gpg-sign", "--message", "Dummy", fileName)
	Expect(err).ToNot(HaveOccurred())
}

func createGitDirWithTag(tag string) {
	_, err := git("init")
	Expect(err).ToNot(HaveOccurred())

	createCommit(tag)

	_, err = git("tag", tag)
	Expect(err).ToNot(HaveOccurred())
}

func createUncomittedChanges() {
	err := ioutil.WriteFile("tracked_file", []byte("Dummy content"), 0600)
	Expect(err).ToNot(HaveOccurred())

	_, err = git("add", "tracked_file")
	Expect(err).ToNot(HaveOccurred())
}
