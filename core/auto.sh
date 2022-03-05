#!/bin/bash

set -ex

cur=$(dirname $0)
idlc_bin="$cur/ovs/ovsdb/ovsdb-idlc.in"

PYTHONPATH="$cur/ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes python $idlc_bin annotate idlc/confd.ovsschema idlc/confd-idl.ann > idlc/confd-idl.ovsidl
PYTHONPATH="$cur/ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes python $idlc_bin c-idl-source idlc/confd-idl.ovsidl > idlc/confd-idl.c
PYTHONPATH="$cur/ovs/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes python $idlc_bin c-idl-header idlc/confd-idl.ovsidl > idlc/confd-idl.h
