TEMPLATE = app
TARGET = test_ui
CONFIG += warn_on qmltestcase

SOURCES += \
    tests/test_ui.cpp \
    gui/qjsonmodel.cpp \
    gui/handlers.cpp

HEADERS += \
    lib/libgoshim.h \
    gui/qjsonmodel.h \
    gui/handlers.h

LIBS += -L../lib -lgoshim -lpthread

RESOURCES += tests/tests.qrc

DESTDIR = build
OBJECTS_DIR = build/.obj
RCC_DIR = build/.rcc
UI_DIR = build/.ui

Release:DESTDIR = build
Release:OBJECTS_DIR = build/.obj
Release:RCC_DIR = build/.rcc
Release:UI_DIR = build/.ui
