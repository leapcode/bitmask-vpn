@ECHO OFF

IF "%GOPATH%"=="" GOTO NOGO
IF NOT EXIST %GOPATH%\bin\2goarray.exe GOTO INSTALL
:POSTINSTALL
call :create_go on On
call :create_go off Off
call :create_go blocked Blocked
call :create_go wait_0 Wait0
call :create_go wait_1 Wait1
call :create_go wait_2 Wait2
call :create_go wait_3 Wait3
GOTO DONE

:create_go
ECHO Creating %1_win.go
ECHO //+build windows > %1_win.go
ECHO. >> %1_win.go
TYPE ico\white\vpn_%1.ico | %GOPATH%\bin\2goarray %2 icon >> %1_win.go
EXIT /B

:CREATEFAIL
ECHO Unable to create output file
GOTO DONE

:INSTALL
ECHO Installing 2goarray...
go get github.com/cratonica/2goarray
IF ERRORLEVEL 1 GOTO GETFAIL
GOTO POSTINSTALL

:GETFAIL
ECHO Failure running go get github.com/cratonica/2goarray.  Ensure that go and git are in PATH
GOTO DONE

:NOGO
ECHO GOPATH environment variable not set
GOTO DONE

:DONE

