package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/heroku/docker-registry-client/registry"
	"github.com/pkg/errors"
)

func imageTags(serverURL, username, password, image string) ([]string, error) {
	log.SetOutput(ioutil.Discard)

	hub, err := registry.New(serverURL, username, password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to registry")
	}

	tags, err := hub.Tags(image)
	if err != nil {
		return []string{}, nil
	}

	return tags, nil
}

func calculateTagsForVersion(version string, tags []string) ([]string, error) {
	major, minor, _, tail, err := parts(version)
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
