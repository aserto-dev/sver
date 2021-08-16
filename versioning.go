package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	gitBinary = "git"

	// Based on https://semver.org/#semantic-versioning-200 but we do support the
	// common `v` prefix in front and do not allow plus elements like `1.0.0+gold`.
	regexSupportedVersionFormat = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?$`)

	regexMajor = regexp.MustCompile(`^([0-9]+)\.[0-9]+\.[0-9]+.*`)
	regexMinor = regexp.MustCompile(`^[0-9]+\.([0-9]+)\.[0-9]+.*`)
	regexPatch = regexp.MustCompile(`^[0-9]+\.[0-9]+\.([0-9]+).*`)
	regexTail  = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(.*)`)
)

func currentVersion() (string, error) {
	err := verifyGit()
	if err != nil {
		return "", errors.Wrap(err, "git error")
	}

	hasTag := true
	tag, err := git("describe", "--tags", "--abbrev=0")
	if err != nil {
		if !strings.Contains(err.Error(), "cannot describe anything") {
			return "", errors.Wrap(err, "exec error")
		} else {
			tag = "0.0.0"
			hasTag = false
		}
	}

	if !regexSupportedVersionFormat.MatchString(tag) {
		if strings.Contains(tag, "+") {
			return "", errors.Errorf("looks like your git tag '%s' has a semver with a + sign - that's not supported by this tool", tag)
		}

		return "", errors.Errorf("'%s' doesn't seem to be a semantic version", tag)
	}

	version := tag

	// Version starts being the last tag that points to a commit in the branch,
	// then it gets mutated based on a series of constraints.

	//  If the tag doesn't point to HEAD, it's a pre-release.
	pointsAt, err := git("tag", "--points-at", "HEAD")
	if err != nil {
		return "", errors.Wrap(err, "exec error")
	}
	if pointsAt == "" {
		// The commit timestamp should be in the format yyyymmddHHMMSS in UTC.
		gitCommitTimestamp, err := git("show", "--no-patch", "--format=%ct", "HEAD")
		if err != nil {
			return "", errors.Wrap(err, "exec error")
		}

		unixTime, err := strconv.ParseInt(gitCommitTimestamp, 10, 64)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse git commit timestamp")
		}
		parsedTimestamp := time.Unix(unixTime, 0)
		gitCommitTimestamp = parsedTimestamp.Format("20060102150405")

		//  The number of commits since last tag that points to a commits in the
		//  branch.
		gitNumberCommits := "0"
		if hasTag {
			gitNumberCommits, err = git("rev-list", "--count", fmt.Sprintf("%s...HEAD", version))
		}
		if err != nil {
			return "", errors.Wrap(err, "exec error")
		}

		//  Add `g` to the short hash to match git describe.
		gitCommitShortHash, err := git("rev-parse", "--short=8", "HEAD")
		if err != nil {
			return "", errors.Wrap(err, "exec error")
		}

		gitCommitShortHash = "g" + gitCommitShortHash

		//  The version gets assembled with the pre-release part.
		version = fmt.Sprintf("%s-%s.%s.%s", version, gitCommitTimestamp, gitNumberCommits, gitCommitShortHash)
	}

	// If there's a change in the source tree that didn't get committed, append
	// `-dirty` to the version string.
	status, err := git("status", "--short")
	if err != nil {
		return "", errors.Wrap(err, "exec error")
	}
	if status != "" {
		version = fmt.Sprintf("%s-dirty", version)
	}

	version = strings.TrimPrefix(version, "v")

	return version, nil
}

func preRelease(currentVersion, identifier string) string {
	return fmt.Sprintf("%s-%s", currentVersion, identifier)
}

func next(currentVersion, nextType string) (string, error) {
	major, minor, patch, _, err := parts(currentVersion)
	if err != nil {
		return "", errors.Wrap(err, "failed to get version parts")
	}

	switch nextType {
	case "patch":
		patch = patch + 1
	case "minor":
		minor = minor + 1
		patch = 0
	case "major":
		major = major + 1
		minor = 0
		patch = 0
	default:
		return "", errors.Errorf("Invalid value '%s' for next version. Supported values are 'patch', 'minor' and 'major'", nextType)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

func parts(version string) (major, minor, patch uint64, tail string, err error) {
	matches := regexMajor.FindAllStringSubmatch(version, -1)
	if matches == nil || len(matches) < 1 || len(matches[0]) < 2 {
		return 0, 0, 0, "", errors.Errorf("'%s' doesn't look like a semver", version)
	}

	major, err = strconv.ParseUint(matches[0][1], 10, 64)
	if err != nil {
		return 0, 0, 0, "", errors.Errorf("'%s' major part of version is not a positive integer", matches[1][0])
	}

	matches = regexMinor.FindAllStringSubmatch(version, -1)
	if matches == nil || len(matches) < 1 || len(matches[0]) < 2 {
		return 0, 0, 0, "", errors.Errorf("'%s' doesn't look like a semver", version)
	}

	minor, err = strconv.ParseUint(matches[0][1], 10, 64)
	if err != nil {
		return 0, 0, 0, "", errors.Errorf("'%s' minor part of version is not a positive integer", matches[1][0])
	}

	matches = regexPatch.FindAllStringSubmatch(version, -1)
	if matches == nil || len(matches) < 1 || len(matches[0]) < 2 {
		return 0, 0, 0, "", errors.Errorf("'%s' doesn't look like a semver", version)
	}

	patch, err = strconv.ParseUint(matches[0][1], 10, 64)
	if err != nil {
		return 0, 0, 0, "", errors.Errorf("'%s' patch part of version is not a positive integer", matches[1][0])
	}

	matches = regexTail.FindAllStringSubmatch(version, -1)
	if matches == nil || len(matches) < 1 || len(matches[0]) < 2 {
		return 0, 0, 0, "", errors.Errorf("'%s' doesn't look like a semver", version)
	}

	tail = matches[0][1]

	return
}
