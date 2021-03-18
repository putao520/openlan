#!/bin/bash

set -e

mkdir -p build

curdir=$(pwd)
version=$(cat VERSION)

mkdir -p ~/rpmbuild/SOURCES

# update version
sed -e "s/Version:.*/Version:\ ${version}/" ./packaging/openlan-ctrl.spec.in   > build/openlan-ctrl.spec
sed -e "s/Version:.*/Version:\ ${version}/" ./packaging/openlan-point.spec.in  > build/openlan-point.spec
sed -e "s/Version:.*/Version:\ ${version}/" ./packaging/openlan-proxy.spec.in  > build/openlan-proxy.spec
sed -e "s/Version:.*/Version:\ ${version}/" ./packaging/openlan-switch.spec.in > build/openlan-switch.spec

# link source
rm -rf ~/rpmbuild/SOURCES/openlan-"${version}"
ln -s "${curdir}" ~/rpmbuild/SOURCES/openlan-"${version}"
