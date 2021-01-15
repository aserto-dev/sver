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
		return nil, errors.Wrap(err, "failed to get image tags")
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
	sort.Sort(semver.Collection(vs))

	parsedVersion, err := semver.NewVersion(version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse version")
	}

	result := []string{
		version,
		fmt.Sprintf("%d.%d", major, minor),
		fmt.Sprintf("%d", major),
	}

	var latestExistingTag *semver.Version
	if len(vs) > 0 {
		latestExistingTag = vs[len(vs)-1]
	}
	if latestExistingTag == nil || parsedVersion.Compare(latestExistingTag) > 0 {
		result = append(result, "latest")
	}

	return result, nil
}
