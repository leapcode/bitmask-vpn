TARGET = $$TARGET

QT += quickcontrols2 svg
CONFIG += qt staticlib
CONFIG += c++17 strict_c++
CONFIG += qtquickcompiler

RELEASE = $$RELEASE

equals(RELEASE, "yes") {
    message("[qmake] doing release build")
    CONFIG += release
    # debug_and_release is default on windows
    # and needs to be explicitly disabled
    win32:CONFIG -= debug_and_release
} else {
    message("[qmake] doing debug build")
    CONFIG += force_debug_info
    CONFIG += debug_and_release
}

windows:CONFIG -= console
lessThan(QT_MAJOR_VERSION, 5): error("requires Qt 5")
QMAKE_MACOSX_DEPLOYMENT_TARGET = 12
QMAKE_TARGET_BUNDLE_PREFIX = se.leap
QMAKE_BUNDLE = $$TARGET

# The following define makes your compiler emit warnings if you use
# any feature of Qt which as been marked deprecated (the exact warnings
# depend on your compiler). Please consult the documentation of the
# deprecated API in order to know how to port your code away from it.
DEFINES += QT_DEPRECATED_WARNINGS

!defined(VENDOR_PATH, var):VENDOR_PATH="providers/riseup"

message("[qmake] VENDOR_PATH: $$VENDOR_PATH")

RESOURCES += gui/gui.qrc
RESOURCES += $$VENDOR_PATH/vendor.qrc

ICON = $$VENDOR_PATH/icon.png

macx {
    ICON = $$VENDOR_PATH/assets/icon.icns
    LIBS += -framework Security -framework CoreFoundation -lresolv
}
win32 {
    RC_ICONS = $$VENDOR_PATH/assets/icon.ico
}

QT += qml widgets quick

SOURCES += \
    gui/main.cpp \
    gui/qjsonmodel.cpp \
    gui/handlers.cpp


HEADERS += \
    gui/handlers.h \
    gui/qjsonmodel.h \
    lib/libgoshim.h \
    gui/appsettings.h

# we build from build/qt
LIBS += -L../../lib -lgoshim -lpthread

DESTDIR = release
OBJECTS_DIR = release/.obj
MOC_DIR = release/.moc
RCC_DIR = release/.rcc
UI_DIR = release/.ui

Release:DESTDIR = release
Release:OBJECTS_DIR = release/.obj
Release:MOC_DIR = release/.moc
Release:RCC_DIR = release/.rcc
Release:UI_DIR = release/.ui

Debug:DESTDIR = debug
Debug:OBJECTS_DIR = debug/.obj
Debug:MOC_DIR = debug/.moc
Debug:RCC_DIR = debug/.rcc
Debug:UI_DIR = debug/.ui

DISTFILES += \
    README.md

CONFIG += lrelease embed_translations

TRANSLATIONS += $$files(gui/i18n/*.ts, true)
RESOURCES += $$files(gui/i18n/*.qm, true)

# see https://stackoverflow.com/questions/5960192/qml-qt-openurlexternally#5960581
# Needed for bringing browser from background to foreground using
# QDesktopServices: https://bugreports.qt.io/browse/QTBUG-8336
TARGET.CAPABILITY += SwEvent
