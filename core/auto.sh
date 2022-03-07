#!/bin/bash

set -ex

version=$(cat VERSION)
cd $(dirname $0)

build_openvswitch() {
  obj_dir=$(pwd)/../build/obj
  cd ovs && {
    [ -e './configure' ] || ./boot.sh
    [ -e './Makefile' ] || ./configure --prefix=/usr --sysconfdir=/etc --localstatedir=/var
    make -j4 && make install DESTDIR=$obj_dir
    cd -
  }
}

python_bin=python
type $python_bin || python_bin="python3"

build_idlc() {
  idlc_bin="ovs/ovsdb/ovsdb-idlc.in"
  [ -e "idlc/confd.ovsschema" ] || ln -s -f ../../dist/resource/confd.schema.json idlc/confd.ovsschema
  PYTHONPATH="ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes $python_bin $idlc_bin annotate idlc/confd.ovsschema idlc/confd-idl.ann > idlc/confd-idl.ovsidl
  PYTHONPATH="ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes $python_bin $idlc_bin c-idl-source idlc/confd-idl.ovsidl > idlc/confd-idl.c
  PYTHONPATH="ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes $python_bin $idlc_bin c-idl-header idlc/confd-idl.ovsidl > idlc/confd-idl.h
}

update_version() {
  sed -i  "s/#define CORE_PACKAGE_VERSION .*/#define CORE_PACKAGE_VERSION \"$version\"/g" ./version.h
}

update_version
build_openvswitch
build_idlc
