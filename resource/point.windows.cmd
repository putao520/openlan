@ECHO off

SET "OPR=%1%"

IF "%OPR%"=="before" (
	ECHO do something before running
)


SET "DEV=%2%"
SET "ADDR=%3%"

IF "%OPR%"=="after" (
	ECHO update route after running
	netsh interface ipv4 set address %DEV% static %ADDR%
)


IF "%OPR%"=="exit" (
	ECHO clear route on exited
)

