#!/bin/bash

if [ "$1" == "before" ]; then
  echo "do before running"
fi

if [ "$1" == "after" ]; then
  echo "do after running"
fi

if [ "$1" == "exit" ]; then
  echo "do on exit"
fi
