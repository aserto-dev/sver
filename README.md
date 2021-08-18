# sver

[![Go Reference](https://pkg.go.dev/badge/github.com/aserto-dev/sver.svg)](https://pkg.go.dev/github.com/aserto-dev/sver)
[![Coverage Status](https://coveralls.io/repos/github/aserto-dev/sver/badge.svg?branch=main)](https://coveralls.io/github/aserto-dev/sver?branch=main)
[![ci](https://github.com/aserto-dev/sver/actions/workflows/ci.yaml/badge.svg)](https://github.com/aserto-dev/sver/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/aserto-dev/sver)](https://goreportcard.com/report/github.com/aserto-dev/sver)

Calculates semantic versions in a git repo.

`sver` requires the latest tag of the current branch to be a semantic version, otherwise it fails.
Semantic versions with plus elements like `1.0.2+gold` are not supported and also fail.

If the latest tag in the current branch points to `HEAD`, no pre-release version information is added. 

Otherwise, `<commit_timestamp>.<branch_commit_count>.g<commit_short_hash>` is appended to the version string.
`<commit_timestamp>` is in the format `yyyymmddHHMMSS`.

If there are uncommitted changes to the source tree, the `-dirty` string is appended to the final version string.

For example: `1.2.0-20201027184820.3186.g4fc2e9e5-dirty`

## Calculating the next version

`sver` can also calculate the next semantic version based on the current version. To do so, use the `--next` flag. Possible values are `major`, `minor` or `patch`.

## Container image tags

When using the `tags` sub-command, `sver` will look at existing tags in an image repository, and figure out which tags you should apply to the image you're building from the current git commit.

For example if you have tags `1.0.0`, `1.1.0` and `2.1.0` and your current tag is `2.1.0`, running `sver tags` will output:

- `2.1.0` - your new version
- `2.1` - because it’s the latest in the 2.1.* series
- `2` - because it's the latest in the 2.* series
- `latest` - because there’s no other version that’s higher

If you then work on top of tag `1.0.0`, and create a patch tagged `1.0.1`, running `sver tags` will output:

- `1.0.1` - your new version
- `1.0` - because it's the latest in the 1.0.* series

> It’s not outputting `latest` because `2.1.0` is higher than `1.0.1`, and it’s not outputting `1` because `1.1.0` is higher than `1.0.1`.

## See also

The [sver github action](https://github.com/marketplace/actions/sver-semantic-version-calculator).

### Similar projects

- https://github.com/caarlos0/svu
- https://github.com/mdomke/git-semver
- https://github.com/semantic-release/semantic-release
- https://goreleaser.com/

## Credits

Based on [this implementation](https://github.com/cloudfoundry-incubator/kubecf-tools/tree/main/versioning).
