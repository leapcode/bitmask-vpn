import QtQuick
import QtQuick.Controls
import Qt.labs.platform as Labs

Labs.SystemTrayIcon {

    visible: systrayVisible
    property alias statusItem: statusItem

    menu: Labs.Menu {

        id: systrayMenu

        Labs.MenuItem {
            id: statusItem
            text: qsTr("Checking status…")
            enabled: false
        }

        Labs.MenuItem {
            id: vpnSystrayToggle
            text: getConnectionText()
            enabled: isConnectionTextEnabled()
            onTriggered: {
                if (ctx.status == "off") {
                    backend.switchOn()
                } else if (ctx.status == "on") {
                    backend.switchOff()
                }
            }
        }

        Labs.MenuSeparator {}

        Labs.MenuItem {
            text: qsTr("Donate")
            onTriggered: Qt.openUrlExternally(ctx.donateURL)
        }

        Labs.MenuSeparator {}

        Labs.MenuItem {
            id: showAppItem
            //: Part of the systray menu; show or hide the main app window
            text: isVisible() ? qsTr("Hide") : qsTr("Show")
            onTriggered: {
                if (isVisible()) {
                    root.hide()
                } else {
                    root.bringToFront()
                }
            }
        }

        Labs.MenuItem {
            //: Part of the systray menu; quits the application
            text: qsTr("Quit")
            onTriggered: {
                backend.quit()
            }
        }
    }

    function isVisible() {
        return root.visibility != 0 && root.visibility != 3
    }

    function getConnectionText() {
        if (!ctx) {
            return ""
        } else if (ctx.status == "off") {
            // Not Turn on, because we will can later append "to <Location>"
            if (ctx.locations && ctx.bestLocation) {
                return qsTr("Connect to") + " " + getCanonicalLocation(ctx.bestLocation)
            } else {
                return qsTr("Connect")
            }
        } else if (ctx.status == "on") {
            return qsTr("Disconnect")
        } 
        return ""
    }

    function isConnectionTextEnabled() {
        if (!ctx) {
            return false
        }
        return ctx.status == "off" || ctx.status == "on"
    }

    // returns the composite of Location, CC
    function getCanonicalLocation(label) {
        try {
            let loc = ctx.locationLabels[label]
            return loc[0] + ", " + loc[1]
        } catch(e) {
            return "unknown"
        }
    }
}
