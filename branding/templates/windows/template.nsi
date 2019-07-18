SetCompressor /SOLID lzma

!define PRODUCT_PUBLISHER "LEAP Encryption Access Project"
!include "MUI2.nsh"

Name "$applicationName"
Outfile "..\dist\$applicationName-$version.exe"
;TODO make the installdir configurable - and set it in the registry.
InstallDir "C:\Program Files\$applicationName\"
RequestExecutionLevel admin

!include FileFunc.nsh
!insertmacro GetParameters
!insertmacro GetOptions

;Macros

!macro SelectByParameter SECT PARAMETER DEFAULT
	${GetOptions} $R0 "/${PARAMETER}=" $0
	${If} ${DEFAULT} == 0
		${If} $0 == 1
			!insertmacro SelectSection ${SECT}
		${EndIf}
	${Else}
		${If} $0 != 0
			!insertmacro SelectSection ${SECT}
		${EndIf}
	${EndIf}
!macroend



!define BITMAP_FILE riseupvpn.bmp

!define MUI_ICON "..\assets\$applicationNameLower.ico"
!define MUI_UNICON "..\assets\$applicationNameLower.ico"

!define MUI_WELCOMEPAGE_TITLE "$applicationName"
!define MUI_WELCOMEPAGE_TEXT "This will install $applicationName in your computer. $applicationName is a simple, fast and secure VPN Client, powered by Bitmask. \n This VPN service is run by donations from people like you."
#!define MUI_WELCOMEFINISHPAGE_BITMAP "riseup.png"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
 
 

Section "InstallFiles"
  ; first we try to delete the systray, locked by the app.
  ClearErrors
  Delete 'C:\Program Files\$applicationName\bitmask-vpn.exe'
  IfErrors 0 noError

  ; Error handling
  MessageBox MB_OK|MB_ICONEXCLAMATION "$applicationName is Running. Please close it, and then run this installer again."
  Abort

  noError:
  ExecShellWait "runas" "$INSTDIR\nssm.exe" 'stop $applicationNameLower-helper'
  ExecShellWait "runas" "$INSTDIR\nssm.exe" 'remove $applicationNameLower-helper confirm'

  SetOutPath $INSTDIR 
  WriteUninstaller $INSTDIR\uninstall.exe

  ; Add ourselves to Add/remove programs
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower" "DisplayName" "$applicationName"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower" "InstallLocation" "$INSTDIR"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower" "DisplayIcon" "$INSTDIR\icon.ico"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower" "Readme" "$INSTDIR\readme.txt"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower" "DisplayVersion" "$version"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower" "Publisher" "LEAP Encryption Access Project"
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower" "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower" "NoRepair" 1

  ;Start Menu
  createDirectory "$SMPROGRAMS\$applicationName\"
  createShortCut "$SMPROGRAMS\$applicationName\$applicationName.lnk" "$INSTDIR\bitmask-vpn.exe" "" "$INSTDIR\icon.ico"

  File "readme.txt"
  File "..\staging\nssm.exe"
  File "/oname=icon.ico" "..\assets\$applicationNameLower.ico"

  $extra_install_files

SectionEnd

Section "InstallService"
  ; Easy service management thanks to nssm
  ExecWait '"$INSTDIR\nssm.exe" install $applicationNameLower-helper "$INSTDIR\bitmask-helper.exe"'
  ExecWait '"$INSTDIR\nssm.exe" set $applicationNameLower-helper AppDirectory "$INSTDIR"'
  ExecWait '"$INSTDIR\nssm.exe" start $applicationNameLower-helper'
SectionEnd

Section /o "TAP Virtual Ethernet Adapter" SecTAP
	; Adapted from the windows nsis installer from OpenVPN (openvpn-build repo).
	DetailPrint "Installing TAP (may need confirmation)..."
	; The /S flag make it "silent", remove it if you want it explicit
  	ExecWait '"$INSTDIR\tap-windows.exe" /S /SELECT_UTILITIES=1'
	Pop $R0 # return value/error/timeout
	WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\$applicationName" "tap" "installed"
	DetailPrint "TAP installed!"
SectionEnd

Section "Uninstall"
  ExecShellWait "runas" "$INSTDIR\nssm.exe" 'stop $applicationNameLower-helper'
  ExecShellWait "runas" "$INSTDIR\nssm.exe" 'remove $applicationNameLower-helper confirm'

  Delete $INSTDIR\readme.txt
  Delete $INSTDIR\nssm.exe
  Delete $INSTDIR\helper.log
  Delete "$SMPROGRAMS\$applicationName\$applicationName.lnk"
  RMDir "$SMPROGRAMS\$applicationName\"

  $extra_uninstall_files

  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$applicationNameLower"
  ; uninstaller must be always the last thing to remove
  Delete $INSTDIR\uninstall.exe
  RMDir $INSTDIR
SectionEnd

Function .onInit
	!insertmacro SelectByParameter ${SecTAP} SELECT_TAP 1
FunctionEnd

;----------------------------------------
;Languages
 
!insertmacro MUI_LANGUAGE "English"
