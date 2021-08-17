package sver

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

func ImageTags(repoName, username, password string) ([]string, error) {
	repo, err := name.NewRepository(repoName)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid repo name [%s]", repoName)
	}

	tags, err := remote.List(repo, remote.WithAuth(&authn.Basic{
		Username: username,
		Password: password,
	}))
	if err != nil {
		if tErr, ok := err.(*transport.Error); ok {
			switch tErr.StatusCode {
			case http.StatusUnauthorized:
				return nil, errors.Wrap(err, "authentication to docker registry failed")
			case http.StatusNotFound:
				return []string{}, nil
			}
		}

		return nil, errors.Wrap(err, "failed to list tags from registry")
	}

	return tags, nil
}

func CalculateTagsForVersion(version string, tags []string) ([]string, error) {
	major, minor, _, tail, err := Parts(version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse version")
	}

	if tail != "" {
		return []string{version}, nil
	}

	vs := []*semver.Version{}
	for _, r := range tags {
		v, err := semver.NewVersion(r)
		if err != nil {
			continue
		}

		vs = append(vs, v)
	}

	result := []string{version}
	sort.Sort(semver.Collection(vs))

	parsedVersion, err := semver.NewVersion(version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse version")
	}

	doMajor := true
	doMinor := true
	for idx := range vs {
		if parsedVersion.GreaterThan(vs[idx]) {
			continue
		}

		if vs[idx].Major() == parsedVersion.Major() {
			doMajor = false
		}

		if vs[idx].Minor() == parsedVersion.Minor() {
			doMinor = false
		}

		break
	}

	if doMinor {
		result = append(result, fmt.Sprintf("%d.%d", major, minor))
	}

	if doMajor {
		result = append(result, fmt.Sprintf("%d", major))
	}

	var latestExistingTag *semver.Version
	if len(vs) > 0 {
		latestExistingTag = vs[len(vs)-1]
	}
	if latestExistingTag == nil || parsedVersion.GreaterThan(latestExistingTag) {
		result = append(result, "latest")
	}

	return result, nil
}
