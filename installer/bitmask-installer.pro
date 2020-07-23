TEMPLATE = aux

CONFIG -= debug_and_release

INSTALLER = RiseupVPN-Installer

INPUT = $$PWD/config/config.xml $$PWD/packages
inst.input = INPUT
inst.output = $$INSTALLER
inst.commands = binarycreator -c $$PWD/config/config.xml -p $$PWD/packages ${QMAKE_FILE_OUT}
inst.CONFIG += target_predeps no_link combine

QMAKE_EXTRA_COMPILERS += inst

OTHER_FILES += \
# watch out... it chokes with dashes in the path
    packages/riseupvpn/meta/package.xml \
    packages/riseupvpn/meta/install.js \
    packages/riseupvpn/data/README.txt \

macx {
    OTHER_FILES += "packages/riseupvpn/data/riseup-vpn.app"
    OTHER_FILES += "packages/riseupvpn/data/bitmask-helper"
    OTHER_FILES += "packages/riseupvpn/data/installer.py"
}
linux {
    OTHER_FILES += "packages/riseupvpn/data/riseup-vpn"
    OTHER_FILES += "packages/riseupvpn/data/bitmask-helper"
}
win32{
    OTHER_FILES += "packages/riseupvpn/data/riseup-vpn.exe"
    OTHER_FILES += "packages/riseupvpn/data/helper.exe"
}	
