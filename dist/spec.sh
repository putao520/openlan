#!/bin/bash

set -e

mkdir -p build

version=$(cat VERSION)
package=openlan-switch-$version

mkdir -p ~/rpmbuild/SOURCES

# update version
sed -e "s/Version:.*/Version:\ ${version}/" dist/openlan-switch.spec.in > build/openlan-switch.spec

# build dist.tar
rsync -r --exclude build . build/$package
cd build && {
  tar cf ~/rpmbuild/SOURCES/$package-source.tar $package
  rm -rf $package
  cd -
}
