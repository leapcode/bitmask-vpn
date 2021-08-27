import QtQuick 2.9
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.4

import "logic.js" as Logic

Item {

    anchors.centerIn: parent
    width: parent.width
    property alias displayReconnect: bridgeReconnect.visible
    // TODO get obfs4Available from backend, in case provider doesn't have it
    visible: true

    Column {

        anchors.centerIn: parent
        spacing: 10
        width: parent.width

        CheckBox {
            id: bridgeCheck
            checked: false
            text: qsTr("Use obfs4 bridges")
            font.pixelSize: 14
            anchors.horizontalCenter: parent.horizontalCenter
            onClicked: {
                if (checked) {
                    Logic.setNeedsReconnect(true)
                    bridgeReconnect.visible = true
                    app.useBridges(true)
                } else {
                    // This would also need a "needs reconnect" for de-selecting bridges the next time.
                    // better to wait and see the new connection widgets though
                    Logic.setNeedsReconnect(false)
                    bridgeReconnect.visible = false
                    app.useBridges(false)
                }
            }
        }

        Text {
            id: bridgesInfo
            width: 250
            color: "grey"
            text: qsTr("Select a bridge only if you know that you need it to evade censorship in your country or local network.")
            anchors.horizontalCenter: parent.horizontalCenter
            wrapMode: Text.WordWrap
            visible: !bridgeReconnect.visible
        }

        Text {
            id: bridgeReconnect
            width: 250
            font.pixelSize: 12
            color: "red"
            text: qsTr("An obfs4 bridge will be used the next time you connect to the VPN.")
            anchors.horizontalCenter: parent.horizontalCenter
            wrapMode: Text.WordWrap
            visible: false
        }
    }
}
