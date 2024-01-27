import QtQuick
import QtQuick.Controls
import QtQuick.Effects
import "../themes/themes.js" as Theme

Item {
    id: motdBox
    width: parent.width
    property var motdText: ""
    property var motdLink: ""
    property var url: ""
    anchors.horizontalCenter: parent.horizontalCenter

    Rectangle {

        id: labelWrapper
        color: "transparent"
        height: label.paintedHeight + Theme.windowMargin
        width: parent.width
        anchors.verticalCenter: parent.verticalCenter

        Label {
            id: label
            width: labelWrapper.width - Theme.windowMargin
            anchors.centerIn: parent
            text: motdBox.motdText
            horizontalAlignment: Text.AlignHCenter
            wrapMode: Text.Wrap
            font.pixelSize: Theme.fontSizeSmall - 2
            onLinkActivated: Qt.openUrlExternally(link)
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }
         }

        Label {
            id: link
            color: Theme.green
            width: labelWrapper.width - Theme.windowMargin
            anchors.top: label.bottom
            anchors.topMargin: 10
            text: motdBox.motdLink
            horizontalAlignment: Text.AlignHCenter
            wrapMode: Label.Wrap
            font.pixelSize: Theme.fontSizeSmall
            onLinkActivated: Qt.openUrlExternally(link)
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }
         }
    }
}
