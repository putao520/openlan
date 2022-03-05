#!/bin/bash

set -ex

cd $(dirname $0)

build_openvswitch() {
  obj_dir=$(pwd)/../build/obj
  pushd ovs
  [ -e './configure' ] || ./boot.sh
  [ -e './Makefile' ] || ./configure --prefix=/usr --sysconfdir=/etc --localstatedir=/var
  make -j4 && make install DESTDIR=$obj_dir
  popd
}

build_idlc() {
  idlc_bin="ovs/ovsdb/ovsdb-idlc.in"
  [ -e "idlc/confd.ovsschema" ] || ln -s -f ../../dist/resource/confd.schema.json idlc/confd.ovsschema
  PYTHONPATH="ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes python $idlc_bin annotate idlc/confd.ovsschema idlc/confd-idl.ann > idlc/confd-idl.ovsidl
  PYTHONPATH="ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes python $idlc_bin c-idl-source idlc/confd-idl.ovsidl > idlc/confd-idl.c
  PYTHONPATH="ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes python $idlc_bin c-idl-header idlc/confd-idl.ovsidl > idlc/confd-idl.h
}

build_openvswitch
build_idlc
