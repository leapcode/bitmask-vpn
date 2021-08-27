import QtQuick 2.0
import QtQuick.Controls 2.4
import QtQuick.Controls.Material 2.1
import QtQuick.Layouts 1.14

ToolBar {

    Material.background: Material.backgroundColor
    Material.foreground: "black"
    Material.elevation: 0
    visible: stackView.depth > 1 && ctx !== undefined ? false : true

    Item {

        id: footerRow
        width: root.width

        ToolButton {
            id: gwButton
            anchors.verticalCenter: parent.verticalCenter
            anchors.leftMargin: 10
            anchors.left: parent.left
            anchors.verticalCenterOffset: 5
            icon.source: stackView.depth > 1 ? "" : "../resources/globe.svg"
            onClicked: stackView.push("Locations.qml")
        }

        Label {
            id: locationLabel
            anchors.left: gwButton.right
            anchors.verticalCenter: parent.verticalCenter
            anchors.verticalCenterOffset: 5
            text: "Seattle"
        }

        Item {
            Layout.fillWidth: true
            height: gwButton.implicitHeight
        }

        Image {
            id: bridge
            height: 24
            width: 24
            source: "../resources/bridge.png"
            anchors.verticalCenter: parent.verticalCenter
            anchors.verticalCenterOffset: 5
            anchors.right: gwQuality.left
            anchors.rightMargin: 10
        }

        Image {
            id: gwQuality
            height: 24
            width: 24
            source: "../resources/reception-0.svg"
            anchors.right: parent.right
            anchors.rightMargin: 20
            anchors.verticalCenter: parent.verticalCenter
            anchors.verticalCenterOffset: 5
        }
    }
}
