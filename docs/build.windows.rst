windows build notes (still some manual steps needed, this should be further automated).
=======================================================================================

PROVIDER=DemoLib make helper
INSTALLER_DATA=branding/qtinstaller/packages/root.win_x86_64/data/
mkdir -p INSTALLER_DATA
mv main.exe ${INSTALLER_DATA}/helper.exe
TARGET=demolib-vpn make build
TARGET=demolib-vpn make installer_win
