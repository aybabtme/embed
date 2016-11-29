#!/usr/bin/env bash

usage() {
    echo "USAGE: ./release.sh [version] [msg...]"
    exit 1
}

REVISION=$(git rev-parse HEAD)
GIT_TAG=$(git name-rev --tags --name-only $REVISION)
if [ "$GIT_TAG" = "" ]; then
    GIT_TAG="devel"
fi


VERSION=$1
if [ "$VERSION" = "" ]; then
    echo "Need to specify a version! Perhaps '$GIT_TAG'?"
    usage
fi

set -u -e

rm -rf /tmp/embed_build/

mkdir -p /tmp/embed_build/linux
GOOS=linux go build -ldflags "-X main.version=$VERSION" -o /tmp/embed_build/linux/embed ../
pushd /tmp/embed_build/linux/
tar cvzf /tmp/embed_build/embed_linux.tar.gz embed
popd

mkdir -p /tmp/embed_build/darwin
GOOS=darwin go build -ldflags "-X main.version=$VERSION" -o /tmp/embed_build/darwin/embed ../
pushd /tmp/embed_build/darwin/
tar cvzf /tmp/embed_build/embed_darwin.tar.gz embed
popd

temple file < README.tmpl.md > ../README.md -var "version=$VERSION"
git add ../README.md
git commit -m 'release bump'

hub release create \
    -a /tmp/embed_build/embed_linux.tar.gz \
    -a /tmp/embed_build/embed_darwin.tar.gz \
    $VERSION

git push origin master
