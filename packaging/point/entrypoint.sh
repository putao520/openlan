#!/usr/bin/env bash

dir=$(dirname $0)
app=${dir}/point.linux.x86_64

if [ "${VS_ADDR}" == "" ] || [ "${VS_AUTH}" == "" ]; then
  $app -conf ${dir}/point.json
else
  if [ "${VS_TLS}" == "" ]; then
    VS_TLS=false
  fi
  if [ "${LOG_LEVEL}" == "" ]; then
    LOG_LEVEL=2
  fi
    cat > ${dir}/point.new.json <<EOF
{
  "vs.tls": ${VS_TLS},
  "vs.addr": "${VS_ADDR}",
  "vs.auth": "${VS_AUTH}",
  "log.level": ${LOG_LEVEL},
  "log.file": "${dir}/point.log"
}
EOF
  $app -conf ${dir}/point.new.json
fi

