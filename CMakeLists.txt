cmake_minimum_required(VERSION 3.16)
project(bitmask VERSION 1.0 LANGUAGES CXX)

set(CMAKE_INCLUDE_CURRENT_DIR ON)

find_package(QT NAMES Qt5 Qt6 REQUIRED)
find_package(Qt${QT_VERSION_MAJOR} REQUIRED COMPONENTS Core Gui Qml Quick QuickControls2 Svg Widgets LinguistTools)

find_library(LIB_TO_INCLUDE goshim lib)
#qt_standard_project_setup()
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTOUIC ON)

qt_add_executable(bitmask WIN32 MACOSX_BUNDLE
    gui/handlers.cpp gui/handlers.h
    gui/main.cpp
    gui/qjsonmodel.cpp gui/qjsonmodel.h
    lib/libgoshim.h
)
qt_add_qml_module(bitmask
    URI bitmask
    VERSION ${PROJECT_VERSION}
    QML_FILES
        gui/main.qml
        gui/themes/themes.js
        gui/themes/theme-riseup.js
        gui/themes/Riseup.qml
        gui/themes/theme-calyx.js
        gui/themes/Calyx.qml
        gui/themes/ThemeObject.qml
        gui/components/MainView.qml
        gui/components/ThemedPage.qml
        gui/components/Splash.qml
        gui/components/FadeBehavior.qml
        gui/components/Home.qml
        gui/components/Header.qml
        gui/components/Footer.qml
        gui/components/About.qml
        gui/components/MaterialRadioButton.qml
        gui/components/MaterialRadioIndicator.qml
        gui/components/WrappedRadioButton.qml
        gui/components/MaterialCheckBox.qml
        gui/components/StatusBox.qml
        gui/components/Systray.qml
        gui/components/Help.qml
        gui/components/Locations.qml
        gui/components/SignalIcon.qml
        gui/components/Preferences.qml
        gui/components/BoldLabel.qml
        gui/components/LightLabel.qml
        gui/components/VPNButtonBase.qml
        gui/components/VPNMouseArea.qml
        gui/components/VerticalSpacer.qml
        gui/components/Icon.qml
        gui/components/MaterialButton.qml
        gui/components/VPNState.qml
        gui/components/InitErrors.qml
        gui/components/ErrorBox.qml
        gui/components/MotdBox.qml
        gui/components/NavigationDrawer.qml
    RESOURCES
        gui/qtquickcontrols2.conf
        gui/resources/icon-noshield.svg
        gui/resources/location.svg
        gui/resources/settings.svg
        gui/resources/power.svg
        gui/resources/language.svg
        gui/resources/lightning.svg
        gui/resources/close.svg
        gui/resources/donate.svg
        gui/resources/tools.svg
        gui/resources/help.svg
        gui/resources/about.svg
        gui/resources/bridge.svg
        gui/resources/gear-fill.svg
        gui/resources/reception-0.svg
        gui/resources/reception-2.svg
        gui/resources/reception-4.svg
        gui/resources/reception-0@24.svg
        gui/resources/reception-4@24.svg
        gui/resources/arrow-left.svg
        gui/resources/globe.svg
        gui/resources/spy.gif
        gui/resources/quit.svg
        gui/resources/alert.svg
        gui/resources/angle-right.svg
        gui/resources/snowflake.svg
        gui/resources/ravens.svg
        gui/resources/ravens.gif
        gui/resources/riseup-icon.svg
        gui/resources/calyx/calyx-shield-green.png
        gui/resources/calyx/calyx-hand.png
        gui/resources/calyx/calyx-red-shield.png
        gui/resources/calyx/calyx-yellow-shield.png
        gui/resources/fonts/Poppins-Regular.ttf
        gui/resources/fonts/Poppins-Bold.ttf
        gui/resources/fonts/Montserrat-SemiBold.ttf
        gui/resources/fonts/Roboto-Regular.ttf
        gui/resources/fonts/Roboto-Bold.ttf
        gui/assets/icon/png/black/vpn_off.png
        gui/assets/icon/png/black/vpn_on.png
        gui/assets/icon/png/black/vpn_wait_0.png
        gui/assets/icon/png/black/vpn_wait_1.png
        gui/assets/icon/png/black/vpn_wait_2.png
        gui/assets/icon/png/black/vpn_wait_3.png
        gui/assets/icon/png/white/vpn_off.png
        gui/assets/icon/png/white/vpn_on.png
        gui/assets/icon/png/white/vpn_wait_0.png
        gui/assets/icon/png/white/vpn_wait_1.png
        gui/assets/icon/png/white/vpn_wait_2.png
        gui/assets/icon/png/white/vpn_wait_3.png
        gui/assets/img/bird.jpg
        gui/providers/providers.json
    NO_RESOURCE_TARGET_PATH
    OUTPUT_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/release
)
set_source_files_properties("gui/providers/providers.json"
    PROPERTIES QT_RESOURCE_ALIAS "providers.json"
)
set_source_files_properties("gui/resources/fonts/Montserrat-SemiBold.ttf"
    PROPERTIES QT_RESOURCE_ALIAS "monserrat-bold.ttf"
)
set_source_files_properties("gui/resources/fonts/Poppins-Bold.ttf"
    PROPERTIES QT_RESOURCE_ALIAS "poppins-bold.ttf"
)
set_source_files_properties("gui/resources/fonts/Poppins-Regular.ttf"
    PROPERTIES QT_RESOURCE_ALIAS "poppins-regular.ttf"
)
set_source_files_properties("gui/resources/fonts/Roboto-Bold.ttf"
    PROPERTIES QT_RESOURCE_ALIAS "roboto-bold.ttf"
)
set_source_files_properties("gui/resources/fonts/Roboto-Regular.ttf"
    PROPERTIES QT_RESOURCE_ALIAS "roboto.ttf"
)

# Avoid collision between executable name and QML module directory (QTBUG-98568).
if(NOT WIN32 AND NOT APPLE)
    set_property(TARGET bitmask PROPERTY OUTPUT_NAME bitmask_app)
endif()

target_compile_definitions(bitmask PRIVATE
    QT_DEPRECATED_WARNINGS
)

target_link_libraries(bitmask PRIVATE
    # Remove: L../../lib
    Qt::Core
    Qt::Gui
    Qt::Qml
    Qt::Quick
    Qt::QuickControls2
    Qt::Svg
    Qt::Widgets
    ${LIB_TO_INCLUDE}
    pthread
)


# Resources:
set_source_files_properties("providers/riseup/assets/icon.svg"
    PROPERTIES QT_RESOURCE_ALIAS "icon.svg"
)
set(vendor_resource_files
    "providers/riseup/assets/icon.svg"
)

qt_add_resources(bitmask "vendor"
    PREFIX
        "/vendor/"
    BASE
        "providers/riseup"
    FILES
        ${vendor_resource_files}
)

qt_add_translations(bitmask TS_FILES
    gui/i18n/base.ts
    gui/i18n/main_de.ts
    gui/i18n/main_el.ts
    gui/i18n/main_es_ES.ts
    gui/i18n/main_es-rAR.ts
    gui/i18n/main_es.ts
    gui/i18n/main_et.ts
    gui/i18n/main_fa_IR.ts
    gui/i18n/main_fi.ts
    gui/i18n/main_fr.ts
    gui/i18n/main_hu.ts
    gui/i18n/main_it.ts
    gui/i18n/main_lt.ts
    gui/i18n/main_nl.ts
    gui/i18n/main_pl.ts
    gui/i18n/main_pt-rBR.ts
    gui/i18n/main_ru.ts
    gui/i18n/main_th.ts
    gui/i18n/main_tr.ts
    gui/i18n/main_ug.ts
    gui/i18n/main_zh-rTW.ts
    gui/i18n/main_zh.ts
)

if(MACOS)
    target_link_libraries(bitmask PRIVATE
        "-framework CoreFoundation"
        "-framework Security"
    )
endif()

install(TARGETS bitmask
    BUNDLE DESTINATION .
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
)

qt_generate_deploy_qml_app_script(
    TARGET bitmask
    FILENAME_VARIABLE deploy_script
    NO_UNSUPPORTED_PLATFORM_ERROR
    DEPLOY_USER_QML_MODULES_ON_UNSUPPORTED_PLATFORM
    MACOS_BUNDLE_POST_BUILD
)
install(SCRIPT ${deploy_script})
