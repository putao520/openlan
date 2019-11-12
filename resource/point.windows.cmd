@ECHO off

IF "%1%"=="before" (
	ECHO "before running"
)

IF "%1%"=="after" (
	ECHO "after running"
	route ADD 192.168.10.0/24 192.168.4.151
)

IF "%1%"=="exit" (
	ECHO "exited"
	route DELETE 192.168.10.0/24
)