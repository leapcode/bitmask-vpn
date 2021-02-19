!defined(INSTALLER, var):INSTALLER= "BitmaskVPN-Installer-git"
!defined(TARGET, var):TARGET= "bitmask-vpn"
TEMPLATE = aux
CONFIG -= debug_and_release

INPUT = $$PWD/config/config.xml $$PWD/packages
inst.input = INPUT
inst.output = $$INSTALLER
inst.commands = binarycreator --ignore-translations -c $$PWD/config/config.xml -p $$PWD/packages ${QMAKE_FILE_OUT}
inst.CONFIG += target_predeps no_link combine

QMAKE_TARGET_BUNDLE_PREFIX = se.leap
QMAKE_BUNDLE = $$TARGET
QMAKE_EXTRA_COMPILERS += inst

# OTHER_FILES += \

macx {
}
linux {
}
win32{
}	
