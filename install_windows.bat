@echo off

FIND /C /I " go" %WINDIR%\system32\drivers\etc\hosts
IF %ERRORLEVEL% NEQ 0 (
  ECHO "Need to add go to your hosts file."
) ELSE (
  ECHO "Nothing to do."
  EXIT /B 0
)

goto :CheckAdmin
:Modify

FIND /C /I " go" %WINDIR%\system32\drivers\etc\hosts
ECHO %NEWLINE%^127.0.103.111 go>>%WINDIR%\System32\drivers\etc\hosts

ECHO "Done."
ping -n 3 127.0.0.1 >nul
EXIT


:CheckAdmin

:: BatchGotAdmin
:: https://stackoverflow.com/a/20861377/1210278
:-------------------------------------
REM  --> Check for permissions
>nul 2>&1 "%SYSTEMROOT%\system32\cacls.exe" "%WINDIR%\system32\drivers\etc\hosts"

REM --> If error flag set, we do not have admin.
if '%errorlevel%' NEQ '0' (
    echo Requesting administrative privileges...
    goto UACPrompt
) else ( goto Modify )

:UACPrompt
    echo Set UAC = CreateObject^("Shell.Application"^) > "%temp%\getadmin.vbs"
    set params = %*:"="
    echo UAC.ShellExecute "cmd.exe", "/c %~s0 %params%", "", "runas", 1 >> "%temp%\getadmin.vbs"

    "%temp%\getadmin.vbs"
    del "%temp%\getadmin.vbs"
    exit /B

