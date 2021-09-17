import QtQuick 2.9
import QtQuick.Controls 2.2

ThemedPage {
    title: qsTr("Help")
    property var issueTracker: "https://0xacab.org/leap/bitmask-vpn/issues"

    Column {
        anchors.centerIn: parent
        spacing: 10

        Text {
            font.pixelSize: 14
            anchors.horizontalCenter: parent.horizontalCenter
            text: getDummyLink(qsTr("Troubleshooting and support"))
            onLinkActivated: Qt.openUrlExternally(ctx.helpURL)
        }
        Text {
            font.pixelSize: 14
            anchors.horizontalCenter: parent.horizontalCenter
            text: getDummyLink(qsTr("Report a bug"))
            onLinkActivated: Qt.openUrlExternally(issueTracker)
        }
        Button {
            anchors.horizontalCenter: parent.horizontalCenter
            text: qsTr("Open logs")
        }
    }

    function getDummyLink(text) {
        return "<a href='#'>" + text + "</a>"
    }
}
