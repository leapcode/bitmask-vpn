import QtQuick 2.15
import QtQuick.Controls 2.2

import "../themes/themes.js" as Theme

ThemedPage {
    title: qsTr("Help")
    property var issueTracker: "https://0xacab.org/leap/bitmask-vpn/issues"
    property var uninstall: "https://0xacab.org/leap/bitmask-vpn/-/blob/main/docs/uninstall.md"

    Column {
        anchors.centerIn: parent
        spacing: 10

        Text {
            font.pixelSize: 14
            textFormat: Text.RichText
            color: Theme.green
            anchors.horizontalCenter: parent.horizontalCenter
            text: getDummyLink(qsTr("Troubleshooting and support"))
            onLinkActivated: Qt.openUrlExternally(ctx.helpURL)
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }
        }
        Text {
            font.pixelSize: 14
            textFormat: Text.RichText
            color: Theme.green
            anchors.horizontalCenter: parent.horizontalCenter
            text: getDummyLink(qsTr("Report a bug"))
            onLinkActivated: Qt.openUrlExternally(issueTracker)
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }
        }
        Text {
            font.pixelSize: 14
            textFormat: Text.RichText
            color: Theme.green
            anchors.horizontalCenter: parent.horizontalCenter
            text: getDummyLink(qsTr("How to uninstall"))
            onLinkActivated: Qt.openUrlExternally(uninstall)
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }
        }
        /* XXX needs implementation in the backend
        Button {
            anchors.horizontalCenter: parent.horizontalCenter
            text: qsTr("Open logs")
        }
        */
    }

    function getDummyLink(text) {
        return "<style>a:link {color: '" + Theme.green + "';}</style><a href=\"#\">" + text + "</a>"
    }
}
