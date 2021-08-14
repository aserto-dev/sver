# sver

- [sver](#sver)
  - [Calculating the next version](#calculating-the-next-version)

Calculates semantic versions in a git repo.
Based on [this implementation](https://github.com/cloudfoundry-incubator/kubecf-tools/tree/main/versioning).

`sver` requires the latest tag of the current branch to be a semantic version, otherwise it fails.
Semantic versions with plus elements like `1.0.2+gold` are not supported and also fail.

If the latest tag in the current branch points to HEAD, no pre-release version information is added. Otherwise, <commit_timestamp>.<branch_commit_count>.g<commit_short_hash> is appended to the version string. <commit_timestamp> is in the format yyyymmddHHMMSS.

If there are uncommitted changes to the source tree, the -dirty string is appended to the final version string.

For example: 1.2.0-20201027184820.3186.g4fc2e9e5-dirty.

## Calculating the next version
This script can also calculate the next semantic version based on the current version. To do so, use the --next flag. Possible values are major, minor or patch.

- When major is used, the major part of the current semver is bumped and the minor and patch parts are reset to 0.
- When minor is used, the minor part of the current semver is bumped and the patch part is reset to 0.
- When patch is used, the patch part of the current semver is bumped.
