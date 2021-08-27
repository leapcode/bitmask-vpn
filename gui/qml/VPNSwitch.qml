import QtQuick 2.9
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.4

SwitchDelegate {

    text: qsTr("")
    checked: false
    anchors.horizontalCenter: parent.horizontalCenter

    contentItem: Text {
        rightPadding: vpntoggle.indicator.width + vpntoggle.spacing
        text: vpntoggle.text
        font: vpntoggle.font
        opacity: enabled ? 1.0 : 0.5
        color: vpntoggle.down ? "#17a81a" : "#21be2b"
        elide: Text.ElideRight
        verticalAlignment: Text.AlignVCenter
    }

    indicator: Rectangle {
        implicitWidth: 48
        implicitHeight: 26
        x: vpntoggle.width - width - vpntoggle.rightPadding
        y: parent.height / 2 - height / 2
        radius: 13
        color: vpntoggle.checked ? "#17a81a" : "transparent"
        border.color: vpntoggle.checked ? "#17a81a" : "#cccccc"

        Rectangle {
            x: vpntoggle.checked ? parent.width - width : 0
            width: 26
            height: 26
            radius: 13
            color: vpntoggle.down ? "#cccccc" : "#ffffff"
            border.color: vpntoggle.checked ? (vpntoggle.down ? "#17a81a" : "#21be2b") : "#999999"
        }
    }

    background: Rectangle {
        implicitWidth: 100
        implicitHeight: 40
        visible: vpntoggle.down || vpntoggle.highlighted
        color: vpntoggle.down ? "#17a81a" : "#eeeeee"
    }
} // end switchdelegate
