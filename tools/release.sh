#!/bin/bash

# Get the highest tag number
VERSION="$(git describe --abbrev=0 --tags)"
VERSION=${VERSION:-'0.0.0'}

# Get number parts
MAJOR="${VERSION%%.*}"
VERSION="${VERSION#*.}"
MINOR="${VERSION%%.*}"
VERSION="${VERSION#*.}"
PATCH="${VERSION%%.*}"
VERSION="${VERSION#*.}"

# Increase version
PATCH=$((PATCH + 1))

TAG="${1}"

if [ "${TAG}" = "" ]; then
  TAG="${MAJOR}.${MINOR}.${PATCH}"
fi

echo "Releasing ${TAG} ..."

if ! [ -f CHANGELOG.md ]; then
  touch CHANGELOG.md && git add CHANGELOG.md
fi
git-chglog --next-tag="${TAG}" --output CHANGELOG.md
git commit -a -m "Update CHANGELOG for ${TAG}"
git tag -a -s -m "Release ${TAG}" "${TAG}"
git push && git push --tags

if [ -f .goreleaser.yml ]; then
  goreleaser release \
    --rm-dist \
    --release-notes <(git-chglog "${TAG}" | tail -n+5)
else
  tea release create --title "${TAG}" --tag "${TAG}" --note "$(git-chglog "${TAG}" | tail -n+5)"
fi
