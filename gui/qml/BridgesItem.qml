import QtQuick 2.9
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.4

import "logic.js" as Logic

Item {

    anchors.centerIn: parent
    width: parent.width
    property alias displayReconnect: bridgeReconnect.visible

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
                    Logic.setNeedsReconnect(true);
                    bridgeReconnect.visible = true;
                } else {
                    // This would also need a "needs reconnect" for de-selecting bridges the next time.
                    // better to wait and see the new connection widgets though
                    Logic.setNeedsReconnect(false);
                    bridgeReconnect.visible = false;
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
        }

        Text {
            id: bridgeReconnect
            width: 250
            font.pixelSize: 12
            color: "red"
            text: qsTr("We will attempt to connect to a bridge the next time you connect to the VPN.")
            anchors.horizontalCenter: parent.horizontalCenter
            wrapMode: Text.WordWrap 
            visible: false;
        }
    }
}
