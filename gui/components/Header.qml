import QtQuick
import QtQuick.Controls
import QtQuick.Dialogs
import QtQuick.Controls.Material

import "../themes/themes.js" as Theme

ToolBar {
    property bool forceInitialVisibility: false
    property bool isSwitchProvider: false
    property bool isInitialSetup: false

    signal backToMainViewRequested()

    visible: stackView.depth > 1 || forceInitialVisibility

    Component.onCompleted: {
        console.log("[Header] Component completed - isInitialSetup:", isInitialSetup, ", forceInitialVisibility:", forceInitialVisibility);
    }
    Material.foreground: Material.Black
    Material.background: customTheme.bgColor
    Material.elevation: 0

    contentHeight: settingsButton.implicitHeight

    ToolButton {
        id: settingsButton
        anchors {
            left: parent.left
            // margin needed at least for the Locations panel
            leftMargin: 5
        }
        font.pixelSize: Qt.application.font.pixelSize * 1.6
        icon.source: "../resources/arrow-left.svg"
        HoverHandler {
            cursorShape: Qt.PointingHandCursor
        }
        onClicked: {
            console.log("[Header] Back button clicked");
            console.log("[Header] stackView.depth:", stackView.depth);
            console.log("[Header] isInitialSetup:", isInitialSetup);
            console.log("[Header] isSwitchProvider:", isSwitchProvider);
            console.log("[Header] parent type:", parent ? parent.toString() : "null");
            console.log("[Header] settingsDrawer defined?:", typeof settingsDrawer !== 'undefined');

            if (isSwitchProvider && stackView.depth >= 4) {
                backToMainViewRequested()
            } else if (stackView.depth > 1) {
                console.log("[Header] Popping stackView");
                stackView.pop()
            } else if (isSwitchProvider && stackView.depth == 1) {
                console.log("[Header] Emitting backToMainViewRequested signal");
                // In initial setup or switch provider mode, request navigation to MainView
                backToMainViewRequested()
            } else if (typeof settingsDrawer !== 'undefined' && settingsDrawer !== null) {
                console.log("[Header] Toggling settingsDrawer");
                settingsDrawer.toggle()
            } else {
                console.log("[Header] No action taken - all conditions false");
            }
        }
    }

    Label {
        text: stackView.currentItem.title
        font.bold: true
        anchors.centerIn: parent
    }
}
