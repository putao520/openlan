@ECHO off

IF "%1%"=="before" (
	ECHO "do something before running"
)

IF "%1%"=="after" (
	ECHO "update route after running"
	route ADD 192.168.10.0/24 192.168.4.151
)

IF "%1%"=="exit" (
	ECHO "clear route on exited"
	route DELETE 192.168.10.0/24
)
