#TARGET = $$BINARY_NAME

CONFIG += qt staticlib
windows:CONFIG += console
unix:DEBUG:CONFIG += debug
lessThan(QT_MAJOR_VERSION, 5): error("requires Qt 5")
QMAKE_MACOSX_DEPLOYMENT_TARGET = 10.11

!defined(VENDOR_PATH, var):VENDOR_PATH="providers/riseup"

RESOURCES += gui/gui.qrc
RESOURCES += $$VENDOR_PATH/vendor.qrc

ICON = $$VENDOR_PATH/icon.png

macx {
    ICON = $$VENDOR_PATH/assets/icon.icns
    LIBS += -framework Security
}
win32 {
    RC_ICONS = $$VENDOR_PATH/assets/icon.ico
}

QT += qml quick widgets

SOURCES += \
    gui/main.cpp \
    gui/qjsonmodel.cpp \
    gui/handlers.cpp


HEADERS += \
    gui/handlers.h \
    gui/qjsonmodel.h \
    lib/libgoshim.h

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
