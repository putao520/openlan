#!/bin/bash

set -e 
set -v

version=$(cat VERSION)
mkdir -p ~/rpmbuild/SOURCES

# update version
sed -i  -e "s/Version:.*/Version:\ ${version}/" ./packaging/openlan-point.spec
sed -i  -e "s/Version:.*/Version:\ ${version}/" ./packaging/openlan-point.spec

# link source
rm -rf ~/rpmbuild/SOURCES/openlan-${version}        
ln -s $(pwd) ~/rpmbuild/SOURCES/openlan-${version}        
