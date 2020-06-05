# FIXME: this should be overwritten by build templates
TARGET=riseup-vpn

CONFIG += qt staticlib
windows:CONFIG += console
unix:DEBUG:CONFIG += debug
lessThan(QT_MAJOR_VERSION, 5): error("requires Qt 5")
QMAKE_MACOSX_DEPLOYMENT_TARGET = 10.14

macx {
    LIBS += -framework Security
}

# trying to optimize size of the static binary.
# probably more can be shaved off with some patience
# You need to recompile your version of Qt to use the libraries you want. The
# information comes from the build configuration of the Qt version that you are
# using. Simply point Qts configure to the relevant libraries you wish to
# override, build it, and use it to build your project. It will automatically
# pull in the newer libraries that you overrode.
# TODO: patch the $(PKG)_BUILD definition in mxe/src/qtbase.mk and shave some options there.
# https://stackoverflow.com/questions/5587141/recommended-flags-for-a-minimalistic-qt-build
# See also: https://qtlite.com/

#QTPLUGIN.imageformats = -
#QTPLUGIN.QTcpServerConnectionFactory =-
#QTPLUGIN.QQmlDebugServerFactory =-
#QTPLUGIN.QWindowsIntegrationPlugin =-
#QTPLUGIN.QQmlDebuggerServiceFactory =-
#QTPLUGIN.QQmlInspectorServiceFactory =-
#QTPLUGIN.QLocalClientConnectionFactory =-
#QTPLUGIN.QDebugMessageServiceFactory =-
#QTPLUGIN.QQmlNativeDebugConnectorFactory =-
#QTPLUGIN.QQmlNativeDebugServiceFactory =-
#QTPLUGIN.QQmlPreviewServiceFactory =-
#QTPLUGIN.QQmlProfilerServiceFactory =-
#QTPLUGIN.QQuickProfilerAdapterFactory =-
#QTPLUGIN.QQmlDebugServerFactory =-
#QTPLUGIN.QTcpServerConnectionFactory =-
#QTPLUGIN.QGenericEnginePlugin =-

QT += qml quick widgets

SOURCES += \
    gui/main.cpp \
    gui/qjsonmodel.cpp \
    gui/handlers.cpp

RESOURCES += gui/gui.qrc

HEADERS += \
    gui/handlers.h \
    gui/qjsonmodel.h \
    lib/libgoshim.h

LIBS += -L../lib -lgoshim -lpthread

DESTDIR = release
OBJECTS_DIR = release/.obj
MOC_DIR = release/.moc
RCC_DIR = release/.rcc
UI_DIR = release/.ui

Release:DESTDIR = release
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
