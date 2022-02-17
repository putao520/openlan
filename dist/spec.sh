#!/bin/bash

set -ex

mkdir -p build

version=$(cat VERSION)
package=openlan-$version

mkdir -p ~/rpmbuild/SOURCES

# update version
sed -e "s/Version:.*/Version:\ ${version}/" dist/openlan.spec.in > build/openlan.spec

# build dist.tar
rsync -r --exclude build . build/$package
cd build && {
  tar cf - $package | gzip -c > ~/rpmbuild/SOURCES/$package-source.tar.gz
  cd -
}
