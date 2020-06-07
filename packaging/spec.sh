#!/bin/bash

set -e

version=$(cat VERSION)
mkdir -p ~/rpmbuild/SOURCES

# update version
sed -i  -e "s/Version:.*/Version:\ ${version}/" ./packaging/openlan-*.spec

# link source
rm -rf ~/rpmbuild/SOURCES/openlan-go-${version}
ln -s $(pwd) ~/rpmbuild/SOURCES/openlan-go-${version}
