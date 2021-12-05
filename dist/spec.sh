#!/bin/bash

set -e

mkdir -p build

curdir=$(pwd)
version=$(cat VERSION)

mkdir -p ~/rpmbuild/SOURCES

# update version
sed -e "s/Version:.*/Version:\ ${version}/" ./dist/openlan-ctrl.spec.in   > build/openlan-ctrl.spec
sed -e "s/Version:.*/Version:\ ${version}/" ./dist/openlan-point.spec.in  > build/openlan-point.spec
sed -e "s/Version:.*/Version:\ ${version}/" ./dist/openlan-proxy.spec.in  > build/openlan-proxy.spec
sed -e "s/Version:.*/Version:\ ${version}/" ./dist/openlan-switch.spec.in > build/openlan-switch.spec

# link source
rm -rf ~/rpmbuild/SOURCES/openlan-*
ln -s "${curdir}" ~/rpmbuild/SOURCES/openlan-"${version}"
