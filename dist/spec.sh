#!/bin/bash

set -e

mkdir -p build

curdir=$(pwd)
version=$(cat VERSION)

mkdir -p ~/rpmbuild/SOURCES

# update version
sed -e "s/Version:.*/Version:\ ${version}/" ./dist/openlan-switch.spec.in > build/openlan-switch.spec

# link source
rm -rf ~/rpmbuild/SOURCES/openlan
cp -rf . ~/rpmbuild/SOURCES/openlan
