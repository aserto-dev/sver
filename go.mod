module github.com/aserto-dev/sver

go 1.14

//replace github.com/aserto-dev/mage-loot => ../mage-loot

require (
	github.com/Masterminds/semver v1.5.0
	github.com/aserto-dev/mage-loot v0.4.16
	github.com/docker/cli v20.10.11+incompatible // indirect
	github.com/docker/docker v20.10.11+incompatible // indirect
	github.com/google/go-containerregistry v0.7.0
	github.com/magefile/mage v1.11.0
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.17.0
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	golang.org/x/sys v0.0.0-20211117180635-dee7805ff2e1 // indirect
)
