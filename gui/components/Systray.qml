import QtQuick 2.0
import QtQuick.Controls 2.4
import Qt.labs.platform 1.0 as Labs

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

        Labs.MenuSeparator {}

        Labs.MenuItem {
            text: qsTr("Donate")
            onTriggered: root.openDonateDialog()
        }

        Labs.MenuSeparator {}

        Labs.MenuItem {
            text: qsTr("Quit")
            onTriggered: backend.quit()
        }
    }
}
