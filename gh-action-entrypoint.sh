#!/usr/bin/env bash
set -e

export PRE_RELEASE="$INPUT_PRE_RELEASE"

if [ -n "${INPUT_DOCKER_IMAGE}" ]; then
  version=$(/app/sver tags -s "${INPUT_DOCKER_REGISTRY}" -u "${INPUT_DOCKER_USERNAME}" -p "${INPUT_DOCKER_PASSWORD}" "${INPUT_DOCKER_IMAGE}")
elif [ -n "${INPUT_NEXT}"]; then
  version=$(/app/sver --next "${INPUT_NEXT}")
else
  version=$(/app/sver)
fi

echo "::set-output name=version::${version//$'\n'/'%0A'}"
