package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	version  = "0.0.0"
	flagNext = ""

	// Based on https://semver.org/#semantic-versioning-200 but we do support the
	// common `v` prefix in front and do not allow plus elements like `1.0.0+gold`.
	regexSupportedVersionFormat = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?$`)

	regexMajor = regexp.MustCompile(`^([0-9]+)\.[0-9]+\.[0-9]+.*`)
	regexMinor = regexp.MustCompile(`^[0-9]+\.([0-9]+)\.[0-9]+.*`)
	regexPatch = regexp.MustCompile(`^[0-9]+\.[0-9]+\.([0-9]+).*`)
	regexTail  = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(.*)`)
)

var rootCmd = &cobra.Command{
	Use: "calc-version [flags]",
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := currentVersion()
		if err != nil {
			return err
		}

		if flagNext != "" {
			version, err = next(version, flagNext)
			if err != nil {
				return err
			}
		}

		fmt.Println(version)

		return nil
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and exit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("calc-version %s\n", version)
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func main() {
	rootCmd.Flags().StringVarP(&flagNext, "next", "n", "", "Prints the next version. Possible values are 'major', 'minor' or 'patch'.")

	rootCmd.AddCommand(
		versionCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func git(args ...string) string {
	err := verifyGit()
	if err != nil {
		fmt.Println(errors.Wrap(err, "git error").Error())
		os.Exit(1)
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(errors.Wrapf(err, "unexpected result from git; output: \n%s\n", string(out)).Error())
		os.Exit(1)
	}

	return strings.TrimSpace(string(out))
}

func currentVersion() (string, error) {
	tag := git("describe", "--tags", "--abbrev=0")

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
	if git("tag", "--points-at", "HEAD") == "" {
		// The commit timestamp should be in the format yyyymmddHHMMSS in UTC.
		gitCommitTimestamp := git("show", "--no-patch", "--format=%ct", "HEAD")
		unixTime, err := strconv.ParseInt(gitCommitTimestamp, 10, 64)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse git commit timestamp")
		}
		parsedTimestamp := time.Unix(unixTime, 0)
		gitCommitTimestamp = parsedTimestamp.Format("20060102150405")

		//  The number of commits since last tag that points to a commits in the
		//  branch.
		gitNumberCommits := git("rev-list", "--count", fmt.Sprintf("%s...HEAD", version))

		//  Add `g` to the short hash to match git describe.
		gitCommitShortHash := git("rev-parse", "--short=8", "HEAD")
		gitCommitShortHash = "g" + gitCommitShortHash

		//  The version gets assembled with the pre-release part.
		version = fmt.Sprintf("%s-%s.%s.%s", version, gitCommitTimestamp, gitNumberCommits, gitCommitShortHash)
	}

	// If there's a change in the source tree that didn't get committed, append
	// `-dirty` to the version string.
	if git("status", "--short") != "" {
		version = fmt.Sprintf("%s-dirty", version)
	}

	version = strings.TrimPrefix(version, "v")

	return version, nil
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
	case "major":
		major = major + 1
	default:
		return "", errors.Errorf("Invalid value '%s' for next version. Supported values are 'patch', 'minor' and 'major'", nextType)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

func verifyGit() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return errors.New("git not found in your PATH; please install it")
	}

	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err = cmd.Run()
	if err != nil {
		return errors.New("current directory is not a git working tree")
	}

	return nil
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
