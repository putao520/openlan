#!/bin/bash

mkdir -p ~/rpmbuild/SOURCES


rm -rf ~/rpmbuild/SOURCES/openlan-${version}
ln -s $(pwd) ~/rpmbuild/SOURCES/openlan-${version}
