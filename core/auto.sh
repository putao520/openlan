#!/bin/bash

set -ex

cd $(dirname $0)

idlc_bin="ovs/ovsdb/ovsdb-idlc.in"

pushd ovs
[ -e './configure' ] || ./boot.sh
[ -e './Makefile' ] || ./configure
make -j4
popd

[ -e "idlc/confd.ovsschema" ] || ln -s -f ../../dist/resource/confd.schema.json idlc/confd.ovsschema

PYTHONPATH="ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes python $idlc_bin annotate idlc/confd.ovsschema idlc/confd-idl.ann > idlc/confd-idl.ovsidl
PYTHONPATH="ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes python $idlc_bin c-idl-source idlc/confd-idl.ovsidl > idlc/confd-idl.c
PYTHONPATH="ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes python $idlc_bin c-idl-header idlc/confd-idl.ovsidl > idlc/confd-idl.h
