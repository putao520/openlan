#!/bin/bash

set -e

curdir=$(pwd)
version=$(cat VERSION)

mkdir -p ~/rpmbuild/SOURCES

# update version

sed -i  -e "s/Version:.*/Version:\ ${version}/" ./packaging/openlan-*.spec

# link source
rm -rf ~/rpmbuild/SOURCES/openlan-"${version}"
ln -s "${curdir}" ~/rpmbuild/SOURCES/openlan-"${version}"
