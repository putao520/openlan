#!/bin/bash

if [ "$1" == "before" ]; then
	echo "do something before running"
fi

if [ "$1" == "after" ]; then
	echo "do something after running"
fi

if [ "$1" == "exit" ]; then
	echo "do something on exit"
fi
