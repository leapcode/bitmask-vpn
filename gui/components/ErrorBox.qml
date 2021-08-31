import QtQuick 2.9
import QtQuick.Controls 2.2
import QtGraphicalEffects 1.0
import "../themes/themes.js" as Theme

Item {
    id: errorBox
    width: parent.width
    property var errorText: ""
    anchors.horizontalCenter: parent.horizontalCenter
    anchors.top: connectionImage.bottom

    // TODO alert icon, by type

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
            text: errorBox.errorText //+ " " + "<b><u>" + alertLinkText + "</b></u>"
            horizontalAlignment: Text.AlignHCenter
            wrapMode: Text.Wrap
            font.pixelSize: Theme.fontSizeSmall
         }
    }
}
