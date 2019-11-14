:: %1%: action
:: %2%: name of device
:: %3%: ipv4 address of device
@ECHO off

SET "OPR=%1%"
SET "DEV=%2%"
SET "ADDR=%3%"

IF "%OPR%"=="before" (
	ECHO do something before running
)

:: route ADD 192.168.10.0/24 192.168.4.151
IF "%OPR%"=="after" (
	ECHO update route after running
	netsh interface ipv4 set address %DEV% static %ADDR%
)

:: route DELETE 192.168.10.0/24
IF "%OPR%"=="exit" (
	ECHO clear route on exited
)
